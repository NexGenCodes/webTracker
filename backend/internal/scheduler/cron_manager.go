package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/usecase"

	"github.com/robfig/cron/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

type CronManager struct {
	scheduler *cron.Cron
	cfg       *config.Config
	shipUC    *usecase.ShipmentUsecase
	configUC  *usecase.ConfigUsecase
	wa        *whatsmeow.Client
	locks     map[string]*sync.Mutex
	mu        sync.RWMutex
}

var (
	instance *CronManager
	once     sync.Once
)

func NewManager(cfg *config.Config, shipUC *usecase.ShipmentUsecase, configUC *usecase.ConfigUsecase, wa *whatsmeow.Client) *CronManager {
	once.Do(func() {
		// Use seconds precision for robfig/cron/v3
		c := cron.New(cron.WithSeconds())
		instance = &CronManager{
			scheduler: c,
			cfg:       cfg,
			shipUC:    shipUC,
			configUC:  configUC,
			wa:        wa,
			locks:     make(map[string]*sync.Mutex),
		}
	})
	return instance
}

func (m *CronManager) Start() {
	// 1. The Pulse: High-frequency logic check (Every minute) for Status Transitions
	// We use a cron job here as a simple ticker wrapper
	m.addJob("The Pulse (Status Updates)", "0 * * * * *", m.handlePulse)

	// 2. Daily Stats Report (At Admin 8 AM - Configured as Cron Spec)
	// For now, hardcode Nigeria 8am approx (07:00 UTC) or use 0 0 8 * * * if system time is local
	// Using "0 0 8 * * *" assuming server TZ is set or we want 8am server time.
	m.addJob("Daily Stats Report", "0 0 8 * * *", m.handleDailyStats)

	// 3. Daily Pruning
	m.addJob("Daily Pruning", "0 0 0 * * *", m.handlePruning)

	// 4. Health Check Heartbeat
	m.addJob("Health Check", "0 */10 * * * *", m.handleHealthCheck)

	m.scheduler.Start()
	logger.Info().Msg("[Cron] Scheduler & Tickers started")
}

func (m *CronManager) Stop() {
	m.scheduler.Stop()
	logger.Info().Msg("[Cron] Native Scheduler stopped")
}

func (m *CronManager) addJob(name, spec string, cmd func()) {
	m.mu.Lock()
	if _, ok := m.locks[name]; !ok {
		m.locks[name] = &sync.Mutex{}
	}
	m.mu.Unlock()

	_, err := m.scheduler.AddFunc(spec, func() {
		m.executeJob(name, cmd)
	})
	if err != nil {
		logger.Error().Err(err).Str("job", name).Msg("Failed to add cron job")
	}
}

func (m *CronManager) executeJob(name string, cmd func()) {
	// Overlap Protection
	m.mu.RLock()
	lock := m.locks[name]
	m.mu.RUnlock()

	if !lock.TryLock() {
		logger.Warn().Str("job", name).Msg("Previous instance still running, skipping")
		return
	}
	defer lock.Unlock()

	logger.Info().Str("job", name).Msg("Starting task")
	cmd()
	logger.Info().Str("job", name).Msg("Task completed")
}

// handlePulse checks for shipments ready to move to next stage (Pending -> Intransit -> Delivered)
func (m *CronManager) handlePulse() {
	ctx := context.Background()
	now := time.Now().UTC()

	// Process transitions in a loop to handle 'catch-up' (cascading statuses)
	maxRounds := 3
	for i := 0; i < maxRounds; i++ {
		transitions, err := m.shipUC.ProcessTransitions(ctx, now)
		if err != nil {
			logger.Error().Err(err).Msg("Pulse: Failed to process status transitions")
			break
		}
		
		if len(transitions) == 0 {
			break
		}

		for _, t := range transitions {
			logger.Info().Str("id", t.TrackingID).Str("new_status", t.NewStatus).Msg("Pulse: Shipment status updated via DB trigger")
			notif.SendStatusAlert(ctx, m.wa, m.cfg, t.UserJID, t.TrackingID, t.NewStatus, t.RecipientEmail)
		}
	}
}

// handleDailyStats compiles and sends the 24h summary
func (m *CronManager) handleDailyStats() {
	ctx := context.Background()

	// Define "Daily" as last 24h
	since := time.Now().Add(-24 * time.Hour)

	created, delivered, err := m.shipUC.CountDailyStats(ctx, since)
	if err != nil {
		logger.Error().Err(err).Msg("Stats: Failed to count")
		return
	}

	msg := fmt.Sprintf("📊 *DAILY REPORT* (Last 24h)\n\n"+
		"📦 *New Shipments:* %d\n"+
		"✅ *Delivered:* %d\n\n"+
		"_System is running smoothly._", created, delivered)

	// Send to Admin/Owner Phone (loaded from config)
	target := m.cfg.BotOwnerPhone
	// Ensure proper JID format for WhatsApp
	if target != "" && !strings.Contains(target, "@") {
		target = target + "@s.whatsapp.net"
	}
	jid, err := types.ParseJID(target)
	if err == nil {
		txt := msg
		m.wa.SendMessage(context.Background(), jid, &waProto.Message{
			Conversation: &txt,
		})
	}
}

// handlePruning removes old data to save space (1GB RAM/Disk constraint)
func (m *CronManager) handlePruning() {
	deliveredCutoff := time.Now().AddDate(0, 0, -7)
	allCutoff := time.Now().AddDate(0, 0, -14)
	deleted, err := m.shipUC.RunAgedCleanup(context.Background(), deliveredCutoff, allCutoff)
	if err != nil {
		logger.Error().Err(err).Msg("Pruning: Failed to run aged cleanup")
		return
	}
	logger.Info().Int64("deleted_count", deleted).Msg("Pruning: Aged cleanup completed successfully")
}

func (m *CronManager) handleHealthCheck() {
	if m.cfg.HealthcheckURL == "" {
		return
	}
	client := &http.Client{Timeout: 5 * time.Second}
	_, err := client.Get(m.cfg.HealthcheckURL)
	if err != nil {
		logger.Error().Err(err).Msg("Health check ping failed")
	}
}

package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/supabase"

	"github.com/robfig/cron/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

type CronManager struct {
	scheduler *cron.Cron
	cfg       *config.Config
	db        *supabase.Client
	wa        *whatsmeow.Client
	locks     map[string]*sync.Mutex
	mu        sync.RWMutex
}

var (
	instance *CronManager
	once     sync.Once
)

func NewManager(cfg *config.Config, db *supabase.Client, wa *whatsmeow.Client) *CronManager {
	once.Do(func() {
		// Use seconds precision for robfig/cron/v3
		c := cron.New(cron.WithSeconds())
		instance = &CronManager{
			scheduler: c,
			cfg:       cfg,
			db:        db,
			wa:        wa,
			locks:     make(map[string]*sync.Mutex),
		}
	})
	return instance
}

func (m *CronManager) Start() {
	// 1. Every 10 Minutes: Native Status Transitions (PENDING -> IN_TRANSIT)
	m.addJob("Status Transitions", "0 */10 * * * *", m.handleTransitions)

	// 2. Every 2 Minutes: Status Change Notifications (Transit Only)
	m.addJob("Status Notifications", "0 */2 * * * *", m.handleNotifications)

	// 3. Every Day at Midnight: Native 7-Day Pruning
	m.addJob("Daily Pruning", "0 0 0 * * *", m.handlePruning)

	// 4. Every 5 Minutes: Health Check Heartbeat
	m.addJob("Health Check", "0 */5 * * * *", m.handleHealthCheck)

	m.scheduler.Start()
	logger.Info().Msg("[Cron] Native Scheduler started successfully")
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

func (m *CronManager) handleTransitions() {
	updated, err := m.db.TransitionPendingToInTransit()
	if err != nil {
		logger.Error().Err(err).Msg("Native transition failed")
		return
	}

	for _, item := range updated {
		m.sendStatusAlert(item.WhatsappFrom, item.TrackingNumber, "IN_TRANSIT")
		_ = m.db.MarkAsNotified(item.TrackingNumber)
	}
}

func (m *CronManager) handleNotifications() {
	jobs, err := m.db.GetPendingNotifications()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to fetch pending notifications")
		return
	}

	for _, j := range jobs {
		m.sendStatusAlert(j.WhatsappFrom, j.TrackingNumber, j.Status)
		_ = m.db.MarkAsNotified(j.TrackingNumber)
		// Small delay to avoid burst
		time.Sleep(500 * time.Millisecond)
	}
}

func (m *CronManager) handlePruning() {
	count, err := m.db.PruneStaleData()
	if err != nil {
		logger.Error().Err(err).Msg("Native pruning failed")
	} else {
		logger.Info().Int("count", count).Msg("Pruned stale records")
	}
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

func (m *CronManager) sendStatusAlert(jidStr, tracking, status string) {
	if jidStr == "" {
		return
	}
	jid, err := types.ParseJID(jidStr)
	if err != nil {
		logger.Warn().Str("jid", jidStr).Msg("Failed to parse JID for status alert")
		return
	}

	var msg string
	switch status {
	case "IN_TRANSIT":
		msg = fmt.Sprintf("🚚 *Status Update*\nID: *%s*\n\nYour package is now *IN TRANSIT*. Our team is handling it at the origin center.", tracking)
	default:
		// Per user request: ONLY notify for transit
		return
	}

	content := &waProto.Message{Conversation: models.StrPtr(msg)}
	_, _ = m.wa.SendMessage(context.Background(), jid, content)
}

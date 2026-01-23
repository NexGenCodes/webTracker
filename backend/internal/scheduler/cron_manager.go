package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/localdb"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/shipment"

	"github.com/robfig/cron/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

type CronManager struct {
	scheduler *cron.Cron
	cfg       *config.Config
	ldb       *localdb.Client // Switch to LocalDB
	wa        *whatsmeow.Client
	locks     map[string]*sync.Mutex
	mu        sync.RWMutex
}

var (
	instance *CronManager
	once     sync.Once
)

func NewManager(cfg *config.Config, ldb *localdb.Client, wa *whatsmeow.Client) *CronManager {
	once.Do(func() {
		// Use seconds precision for robfig/cron/v3
		c := cron.New(cron.WithSeconds())
		instance = &CronManager{
			scheduler: c,
			cfg:       cfg,
			ldb:       ldb,
			wa:        wa,
			locks:     make(map[string]*sync.Mutex),
		}
	})
	return instance
}

func (m *CronManager) Start() {
	// 1. The Pulse: High-frequency logic check (Every 2 minutes) for Status Transitions
	// We use a cron job here as a simple ticker wrapper
	m.addJob("The Pulse (Status Updates)", "0 */2 * * * *", m.handlePulse)

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

	// 1. Check Pending -> Intransit
	pendingIDs, err := m.ldb.GetShipmentsReadyForTransit(ctx)
	if err == nil {
		for _, id := range pendingIDs {
			logger.Info().Str("id", id).Msg("Pulse: Moving shipment to INTRANSIT")
			if err := m.ldb.UpdateShipmentStatus(ctx, id, shipment.StatusIntransit); err != nil {
				logger.Error().Err(err).Str("id", id).Msg("Pulse: Failed to update status")
			} else {
				// Fetch shipment to get JID
				s, _ := m.ldb.GetShipment(ctx, id)
				if s != nil {
					m.sendStatusAlert(s.UserJID, s.TrackingID, shipment.StatusIntransit)
				}
			}
		}
	}

	// 2. Check Intransit -> OutForDelivery
	transitIDs, err := m.ldb.GetShipmentsReadyForOutForDelivery(ctx)
	if err == nil {
		for _, id := range transitIDs {
			logger.Info().Str("id", id).Msg("Pulse: Moving shipment to OUT_FOR_DELIVERY")
			if err := m.ldb.UpdateShipmentStatus(ctx, id, shipment.StatusOutForDelivery); err != nil {
				logger.Error().Err(err).Str("id", id).Msg("Pulse: Failed to update status")
			} else {
				s, _ := m.ldb.GetShipment(ctx, id)
				if s != nil {
					m.sendStatusAlert(s.UserJID, s.TrackingID, shipment.StatusOutForDelivery)
				}
			}
		}
	}

	// 3. Check OutForDelivery -> Delivered
	readyIDs, err := m.ldb.GetShipmentsReadyForDelivery(ctx)
	if err == nil {
		for _, id := range readyIDs {
			logger.Info().Str("id", id).Msg("Pulse: Moving shipment to DELIVERED")
			if err := m.ldb.UpdateShipmentStatus(ctx, id, shipment.StatusDelivered); err != nil {
				logger.Error().Err(err).Str("id", id).Msg("Pulse: Failed to update status")
			} else {
				s, _ := m.ldb.GetShipment(ctx, id)
				if s != nil {
					m.sendStatusAlert(s.UserJID, s.TrackingID, shipment.StatusDelivered)
				}
			}
		}
	}
}

// handleDailyStats compiles and sends the 24h summary
func (m *CronManager) handleDailyStats() {
	ctx := context.Background()

	// Define "Daily" as last 24h
	since := time.Now().Add(-24 * time.Hour)

	created, delivered, err := m.ldb.CountDailyStats(ctx, since)
	if err != nil {
		logger.Error().Err(err).Msg("Stats: Failed to count")
		return
	}

	msg := fmt.Sprintf("📊 *DAILY REPORT* (Last 24h)\n\n"+
		"📦 *New Shipments:* %d\n"+
		"✅ *Delivered:* %d\n\n"+
		"_System is running smoothly._", created, delivered)

	// Send to Admin/Owner Phone (loaded from config)
	// We use the first admin phone or bot owner if configured
	target := m.cfg.BotOwnerPhone
	// If owner phone format is needing parse:
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
	deleted, err := m.ldb.RunAgedCleanup(context.Background())
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

// DELETED DUPLICATE HANDLERS

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
	case shipment.StatusIntransit:
		msg = fmt.Sprintf("🚚 *Status Update*\nID: *%s*\n\nYour package is now *IN TRANSIT*. Our team is handling it at the origin center.", tracking)
	case shipment.StatusOutForDelivery:
		msg = fmt.Sprintf("📦 *Status Update*\nID: *%s*\n\nYour package is *OUT FOR DELIVERY*! Our local agent will contact you shortly.", tracking)
	case shipment.StatusDelivered:
		msg = fmt.Sprintf("✅ *Package Delivered*\nID: *%s*\n\nYour shipment has arrived at the destination. Thank you for choosing our service!", tracking)
	default:
		return
	}

	content := &waProto.Message{Conversation: models.StrPtr(msg)}
	_, _ = m.wa.SendMessage(context.Background(), jid, content)
}

package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/usecase"

	"github.com/robfig/cron/v3"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"webtracker-bot/internal/whatsapp"
)

type CronManager struct {
	scheduler *cron.Cron
	cfg       *config.Config
	shipUC    *usecase.ShipmentUsecase
	configUC  *usecase.ConfigUsecase
	bots      whatsapp.BotProvider
	locks     map[string]*sync.Mutex
	mu        sync.RWMutex
}

var (
	instance *CronManager
	once     sync.Once
)

func NewManager(cfg *config.Config, shipUC *usecase.ShipmentUsecase, configUC *usecase.ConfigUsecase, bots whatsapp.BotProvider) *CronManager {
	once.Do(func() {
		// Use seconds precision for robfig/cron/v3
		c := cron.New(cron.WithSeconds())
		instance = &CronManager{
			scheduler: c,
			cfg:       cfg,
			shipUC:    shipUC,
			configUC:  configUC,
			bots:      bots,
			locks:     make(map[string]*sync.Mutex),
		}
	})
	return instance
}

func (m *CronManager) Start() {
	m.addJob("The Pulse (Status Updates)", "0 * * * * *", m.handlePulse)
	m.addJob("Daily Stats Report", "0 0 8 * * *", m.handleDailyStats)
	m.addJob("Daily Pruning", "0 0 0 * * *", m.handlePruning)
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

func (m *CronManager) handlePulse() {
	ctx := context.Background()
	now := time.Now().UTC()

	companies, err := m.configUC.GetAllCompanies(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Pulse: Failed to get companies")
		return
	}

	for _, companyID := range companies {
		maxRounds := 3
		for i := 0; i < maxRounds; i++ {
			transitions, err := m.shipUC.ProcessTransitions(ctx, companyID, now)
			if err != nil {
				logger.Error().Err(err).Msg("Pulse: Failed to process status transitions")
				break
			}

			if len(transitions) == 0 {
				break
			}

			for _, t := range transitions {
				logger.Info().Str("id", t.TrackingID).Str("new_status", t.NewStatus).Msg("Pulse: Shipment status updated via DB trigger")

				bot, err := m.bots.GetBot(companyID)
				if err != nil {
					continue
				}

				go func(t usecase.TransitionResult, b *whatsapp.BotInstance) {
					notif.SendStatusAlert(ctx, b.WA, m.cfg, b.CompanyName, t.UserJID, t.TrackingID, t.NewStatus, t.RecipientEmail)
				}(t, bot)
			}
		}
	}
}

func (m *CronManager) handleDailyStats() {
	ctx := context.Background()
	since := time.Now().Add(-24 * time.Hour)

	companies, err := m.configUC.GetAllCompanies(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Stats: Failed to get companies")
		return
	}

	for _, companyID := range companies {
		created, delivered, err := m.shipUC.CountDailyStats(ctx, companyID, since)
		if err != nil {
			logger.Error().Err(err).Msg("Stats: Failed to count")
			continue
		}

		msg := fmt.Sprintf("📊 *DAILY STATS*\n\n✅ Created: %d\n📦 Delivered: %d", created, delivered)

		bot, err := m.bots.GetBot(companyID)
		if err != nil {
			continue
		}

		go func() {
			groups, _ := m.configUC.GetAuthorizedGroups(ctx, companyID)
			for _, g := range groups {
				jid, _ := types.ParseJID(g)
				msgContent := &waProto.Message{
					Conversation: &msg,
				}
				bot.WA.SendMessage(ctx, jid, msgContent)
			}
		}()
	}
}

func (m *CronManager) handlePruning() {
	ctx := context.Background()
	deliveredCutoff := time.Now().AddDate(0, 0, -7)
	allCutoff := time.Now().AddDate(0, 0, -14)

	companies, err := m.configUC.GetAllCompanies(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Pruning: Failed to get companies")
		return
	}

	for _, companyID := range companies {
		deleted, err := m.shipUC.RunAgedCleanup(ctx, companyID, deliveredCutoff, allCutoff)
		if err != nil {
			logger.Error().Err(err).Msg("Pruning: Failed to run aged cleanup")
			continue
		}
		logger.Info().Int64("deleted_count", deleted).Msg("Pruning: Aged cleanup completed successfully")
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

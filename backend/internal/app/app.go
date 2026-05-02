package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"database/sql"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/database"
	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/receipt"
	"webtracker-bot/internal/scheduler"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/utils"
	"webtracker-bot/internal/whatsapp"

	transport_http "webtracker-bot/internal/api"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

type App struct {
	Cfg        *config.Config
	ShipmentUC *shipment.Usecase
	ConfigUC   *config.Usecase
	BotManager *whatsapp.Manager
	WAStore    *sqlstore.Container
	WG         sync.WaitGroup
	Cancel     context.CancelFunc
	Cron       *scheduler.CronManager
	Context    context.Context
	SqlPool    *sql.DB
	HttpServer *transport_http.Server
}

func New(cfg *config.Config) *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		Cfg:     cfg,
		Context: ctx,
		Cancel:  cancel,
	}
}

func (a *App) Init() error {
	sqlPool, err := database.Connect(a.Cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database init: %w", err)
	}
	a.SqlPool = sqlPool

	querier := db.New(a.SqlPool)
	shipService := &shipment.Calculator{}
	a.ShipmentUC = shipment.NewUsecase(querier, shipService)
	a.ConfigUC = config.NewUsecase(querier, a.SqlPool)

	dbUrl := a.Cfg.DirectURL
	if dbUrl == "" {
		dbUrl = a.Cfg.DatabaseURL
	}
	store, err := whatsapp.NewStore(dbUrl)
	if err != nil {
		return fmt.Errorf("whatsapp store init: %w", err)
	}
	a.WAStore = store

	a.BotManager = whatsapp.NewManager(a.Context, a.Cfg, a.ShipmentUC, a.ConfigUC, a.WAStore, &a.WG)

	companies, err := a.ConfigUC.GetAllActiveCompanies(context.Background())
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to load active companies")
	}

	for _, c := range companies {
		if err := a.BotManager.InitBotForCompany(c); err != nil {
			logger.Error().Err(err).Str("company", c.Name.String).Msg("Failed to init bot")
		}
	}

	if err := receipt.InitReceiptRenderer(a.Cfg.UseOptimizedReceipt); err != nil {
		logger.Error().Err(err).Msg("Failed to init receipt renderer")
	}

	a.HttpServer = transport_http.NewServer(a.Cfg, a.ShipmentUC, a.ConfigUC, a.SqlPool, a)

	return nil
}

func (a *App) Run() error {
	a.Cron = scheduler.NewManager(a.Cfg, a.ShipmentUC, a.ConfigUC, a)
	a.Cron.Start()

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				utils.CleanupLimits()
			case <-a.Context.Done():
				return
			}
		}
	}()

	for _, bot := range a.BotManager.GetAllBots() {
		wc := bot.GetWAClient()
		if wc != nil && wc.Store != nil && wc.Store.ID != nil {
			if err := wc.Connect(); err != nil {
				logger.Error().Err(err).Str("company", bot.GetCompanyName()).Msg("Failed to connect")
			}
		}
	}

	go func() {
		if err := a.HttpServer.Start(a.Cfg.APIPort); err != nil {
			logger.Error().Err(err).Msg("HTTP Server startup failed")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case sig := <-sigChan:
		logger.Info().Str("signal", sig.String()).Msg("Shutdown signal received")
	case <-a.Context.Done():
		logger.Warn().Msg("App context cancelled (Internal Logout or Error)")
	}

	return a.Shutdown()
}

func (a *App) Shutdown() error {
	logger.Info().Msg("Graceful shutdown initiated...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	a.Cancel()
	var errs []error

	if a.HttpServer != nil {
		if err := a.HttpServer.Stop(); err != nil {
			errs = append(errs, err)
		}
	}

	for _, bot := range a.BotManager.GetAllBots() {
		wc := bot.GetWAClient()
		if wc != nil {
			wc.Disconnect()
		}
		if jobs := bot.GetJobs(); jobs != nil {
			close(jobs)
		}
	}

	if a.Cron != nil {
		a.Cron.Stop()
	}

	receipt.Shutdown()

	done := make(chan struct{})
	go func() {
		a.WG.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info().Msg("All workers stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Warn().Msg("Shutdown timeout reached")
	}

	if a.SqlPool != nil {
		if err := a.SqlPool.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown encountered errors: %v", errs)
	}
	return nil
}

// BotProvider Implementation (Delegation to BotManager)

func (a *App) GetBot(companyID uuid.UUID) (models.BotInstance, error) {
	return a.BotManager.GetBot(companyID)
}

func (a *App) GetAllBots() []models.BotInstance {
	return a.BotManager.GetAllBots()
}

func (a *App) ActivateBot(ctx context.Context, companyID uuid.UUID) error {
	return a.BotManager.ActivateBot(ctx, companyID)
}

func (a *App) DeactivateBot(companyID uuid.UUID) error {
	return a.BotManager.DeactivateBot(companyID)
}

func (a *App) LogoutBot(companyID uuid.UUID) error {
	return a.BotManager.LogoutBot(companyID)
}

func (a *App) PurgeBot(companyID uuid.UUID) error {
	return a.BotManager.PurgeBot(companyID)
}

func (a *App) GeneratePairingCode(ctx context.Context, companyID uuid.UUID, phone string) (string, error) {
	return a.BotManager.GeneratePairingCode(ctx, companyID, phone)
}

func (a *App) GetQR(ctx context.Context, companyID uuid.UUID) (string, error) {
	return a.BotManager.GetQR(ctx, companyID)
}

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
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/receipt"
	"webtracker-bot/internal/scheduler"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/utils"
	"webtracker-bot/internal/whatsapp"
	"webtracker-bot/internal/worker"

	"github.com/google/uuid"

	"webtracker-bot/internal/adapter/db"
	"webtracker-bot/internal/adapter/dbconn"
	transport_http "webtracker-bot/internal/transport/http"
	"webtracker-bot/internal/usecase"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type App struct {
	Cfg        *config.Config
	ShipmentUC *usecase.ShipmentUsecase
	ConfigUC   *usecase.ConfigUsecase
	WAStore    *sqlstore.Container
	Bots       map[uuid.UUID]*whatsapp.BotInstance
	BotsMu     sync.RWMutex
	Jobs       chan models.Job
	WG         sync.WaitGroup
	Cancel     context.CancelFunc
	Cron       *scheduler.CronManager
	Context    context.Context
	SqlPool    *sql.DB
	HttpServer *transport_http.Server
}

func (a *App) GetBot(companyID uuid.UUID) (*whatsapp.BotInstance, error) {
	a.BotsMu.RLock()
	defer a.BotsMu.RUnlock()
	bot, ok := a.Bots[companyID]
	if !ok {
		return nil, fmt.Errorf("bot not found for company %s", companyID)
	}
	return bot, nil
}

// ActivateBot dynamically starts a bot instance for a company.
func (a *App) ActivateBot(ctx context.Context, companyID uuid.UUID) error {
	company, err := a.ConfigUC.GetCompanyByID(ctx, companyID)
	if err != nil {
		return fmt.Errorf("failed to get company: %w", err)
	}

	if !company.AuthStatus.Valid || company.AuthStatus.String != "active" {
		return fmt.Errorf("company is not active")
	}

	a.BotsMu.RLock()
	if _, exists := a.Bots[companyID]; exists {
		a.BotsMu.RUnlock()
		return fmt.Errorf("bot already active for company")
	}
	a.BotsMu.RUnlock()

	return a.initBotForCompany(company)
}

// DeactivateBot dynamically stops and removes a bot instance.
func (a *App) DeactivateBot(companyID uuid.UUID) error {
	a.BotsMu.Lock()
	bot, exists := a.Bots[companyID]
	if !exists {
		a.BotsMu.Unlock()
		return fmt.Errorf("bot not found")
	}

	bot.WA.Disconnect()
	delete(a.Bots, companyID)
	a.BotsMu.Unlock()

	logger.Info().Str("company_id", companyID.String()).Msg("Bot dynamically deactivated")
	return nil
}

func (a *App) GetAllBots() []*whatsapp.BotInstance {
	a.BotsMu.RLock()
	defer a.BotsMu.RUnlock()
	var bots []*whatsapp.BotInstance
	for _, b := range a.Bots {
		bots = append(bots, b)
	}
	return bots
}

func New(cfg *config.Config) *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		Cfg:     cfg,
		Jobs:    make(chan models.Job, cfg.BufferSize),
		Context: ctx,
		Cancel:  cancel,
	}
}

func (a *App) Init() error {
	// Init Clean Architecture dependencies (SQLC + UseCases)
	sqlPool, err := dbconn.Connect(a.Cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("dbconn init (Pgx): %w", err)
	}
	a.SqlPool = sqlPool

	querier := db.New(a.SqlPool)
	shipService := &shipment.Calculator{}
	a.ShipmentUC = usecase.NewShipmentUsecase(querier, shipService)
	a.ConfigUC = usecase.NewConfigUsecase(querier, a.SqlPool)

	// Initialize Multi-Bot Store
	store, err := whatsapp.NewStore(a.Cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("whatsapp store init: %w", err)
	}
	a.WAStore = store
	a.Bots = make(map[uuid.UUID]*whatsapp.BotInstance)

	companies, err := a.ConfigUC.GetAllActiveCompanies(context.Background())
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to load active companies")
	}

	for _, c := range companies {
		if err := a.initBotForCompany(c); err != nil {
			logger.Error().Err(err).Str("company", c.Name).Msg("Failed to init bot")
		}
	}

	if err := utils.InitReceiptRenderer(a.Cfg.UseOptimizedReceipt); err != nil {
		logger.Error().Err(err).Msg("Failed to init receipt renderer")
	}

	// Init Fiber HTTP REST API Server
	// Note: We need to pass the app or a way to get the Sender per company.
	// For now, HttpServer doesn't strictly need a global Sender, it can lookup per company.
	a.HttpServer = transport_http.NewServer(a.Cfg, a.ShipmentUC, a.ConfigUC, a.SqlPool, a)

	return nil
}

func (a *App) initBotForCompany(c db.Company) error {
	var device *store.Device
	var err error

	phone := ""
	if c.WhatsappPhone.Valid {
		phone = c.WhatsappPhone.String
	}

	if phone != "" {
		jid := types.NewJID(phone, "s.whatsapp.net")
		device, err = a.WAStore.GetDevice(context.Background(), jid)
		if err != nil {
			logger.Warn().Err(err).Str("phone", phone).Msg("Device not found for phone, creating new")
			device = a.WAStore.NewDevice()
		}
	} else {
		device = a.WAStore.NewDevice()
	}

	waClient := whatsapp.NewClientForDevice(device)

	// Create prefix from name
	prefix := config.GenerateAbbreviation(c.Name)

	sender := whatsapp.NewSender(waClient, c.Name)
	receipt.InitProcessor()

	bot := &whatsapp.BotInstance{
		CompanyID:   c.ID,
		CompanyName: c.Name,
		Prefix:      prefix,
		Tier:        c.SubscriptionStatus.String,
		WA:          waClient,
		Sender:      sender,
	}

	waClient.AddEventHandler(func(evt interface{}) {
		a.handleWAEvent(bot, evt)
	})

	a.BotsMu.Lock()
	a.Bots[c.ID] = bot
	a.BotsMu.Unlock()

	logger.Info().Str("company", c.Name).Msg("Initialized bot instance")
	return nil
}

func (a *App) Run() error {
	shipmentService := a.ShipmentUC.Service

	// We'll run one worker pool globally that handles jobs from all bots.
	// The Jobs channel carries the CompanyID, so workers know which bot to use.
	for i := 1; i <= a.Cfg.WorkerPoolSize; i++ {
		a.WG.Add(1)
		w := &worker.Worker{
			ID:              i,
			Jobs:            a.Jobs,
			WG:              &a.WG,
			Cfg:             a.Cfg,
			ShipmentUC:      a.ShipmentUC,
			ConfigUC:        a.ConfigUC,
			TrackingBaseURL: a.Cfg.TrackingBaseURL,
			ShipmentService: shipmentService,
			Bots:            a, // 'a' implements BotProvider
		}
		go w.Start()
	}

	// The CronManager needs to loop through bots
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

	if err := a.connectWA(); err != nil {
		return err
	}

	// Start HTTP REST API
	go func() {
		if err := a.HttpServer.Start(a.Cfg.APIPort); err != nil {
			logger.Error().Err(err).Msg("HTTP Server startup failed")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
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

func (a *App) connectWA() error {
	for _, bot := range a.GetAllBots() {
		if bot.WA.Store.ID != nil {
			if err := bot.WA.Connect(); err != nil {
				logger.Error().Err(err).Str("company", bot.CompanyName).Msg("Failed to connect")
			}
		}
	}
	return nil
}

// GeneratePairingCode dynamically connects a bot and returns a pairing code for a given phone number.
func (a *App) GeneratePairingCode(ctx context.Context, companyID uuid.UUID, phone string) (string, error) {
	bot, err := a.GetBot(companyID)
	if err != nil {
		return "", err
	}

	if err := bot.WA.Connect(); err != nil {
		return "", fmt.Errorf("failed to connect to WhatsApp: %w", err)
	}

	code, err := bot.WA.PairPhone(ctx, phone, true, whatsmeow.PairClientChrome, "Chrome (Windows)")
	if err != nil {
		return "", fmt.Errorf("failed to pair phone: %w", err)
	}

	return code, nil
}

func (a *App) handleWAEvent(bot *whatsapp.BotInstance, evt interface{}) {
	whatsapp.HandleEvent(bot.WA, evt, a.Jobs, a.Cfg, a.ConfigUC, bot.CompanyID)

	switch evt.(type) {
	case *events.Connected, *events.PairSuccess:
		logger.Info().Str("company", bot.CompanyID.String()).Msg("WhatsApp Connected/Paired!")
		err := a.ConfigUC.UpdateCompanyAuthStatus(context.Background(), bot.CompanyID, "active")
		if err != nil {
			logger.Error().Err(err).Msg("Failed to update company auth_status to active")
		}
	case *events.LoggedOut:
		logger.Warn().Str("company", bot.CompanyID.String()).Msg("WhatsApp Logged Out!")
		err := a.ConfigUC.UpdateCompanyAuthStatus(context.Background(), bot.CompanyID, "pending")
		if err != nil {
			logger.Error().Err(err).Msg("Failed to update company auth_status to pending")
		}
		a.Cancel()
	}
}

func (a *App) Shutdown() error {
	logger.Info().Msg("Graceful shutdown initiated...")

	// Create a context with a timeout for the shutdown process
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Signal all background goroutines to stop
	a.Cancel()

	var errs []error

	// 1. Stop HTTP Server
	if a.HttpServer != nil {
		if err := a.HttpServer.Stop(); err != nil {
			logger.Error().Err(err).Msg("Error stopping HTTP server")
			errs = append(errs, err)
		}
	}

	// 2. Disconnect WhatsApp
	for _, bot := range a.GetAllBots() {
		bot.WA.Disconnect()
	}
	logger.Info().Msg("All WhatsApp clients disconnected")

	// 3. Stop Scheduler
	if a.Cron != nil {
		a.Cron.Stop()
		logger.Info().Msg("Cron scheduler stopped")
	}

	// 4. Close Receipt Queue (drain goroutines)
	receipt.Shutdown()

	// 5. Close Worker Channel and Wait
	close(a.Jobs)

	// Create a channel to wait for workers
	done := make(chan struct{})
	go func() {
		a.WG.Wait()
		close(done)
	}()

	// Wait for workers or timeout
	select {
	case <-done:
		logger.Info().Msg("All workers stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Warn().Msg("Shutdown timeout reached - some workers may still be running")
	}

	// 5. Close Database Connection
	if a.SqlPool != nil {
		if err := a.SqlPool.Close(); err != nil {
			logger.Error().Err(err).Msg("Error closing database connection")
			errs = append(errs, err)
		} else {
			logger.Info().Msg("Database connection closed")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown encountered errors: %v", errs)
	}

	logger.Info().Msg("App shutdown complete.")
	return nil
}

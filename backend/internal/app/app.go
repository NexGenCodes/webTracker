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
	"webtracker-bot/internal/cache"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/database"
	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/receipt"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/tasks"
	"webtracker-bot/internal/utils"
	"webtracker-bot/internal/whatsapp"
	"webtracker-bot/internal/worker"

	transport_http "webtracker-bot/internal/api"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

type App struct {
	Cfg            *config.Config
	ShipmentUC     *shipment.Usecase
	ConfigUC       *config.Usecase
	BotManager     *whatsapp.Manager
	WAStore        *sqlstore.Container
	WG             sync.WaitGroup
	Cancel         context.CancelFunc
	AsynqScheduler *asynq.Scheduler
	Context        context.Context
	SqlPool        *sql.DB
	HttpServer     *transport_http.Server
	AsynqServer    *asynq.Server
	Worker         *worker.Worker
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

	// Run embedded migrations automatically
	if err := database.RunMigrations(a.Context, a.SqlPool); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	// Seed super admin from env vars (idempotent)
	if err := database.SeedSuperAdmin(a.Context, a.SqlPool, a.Cfg.SuperAdminCompanyEmail, a.Cfg.SuperAdminPassword); err != nil {
		return fmt.Errorf("super admin seed failed: %w", err)
	}

	querier := db.New(a.SqlPool)
	shipService := &shipment.Calculator{}
	a.ShipmentUC = shipment.NewUsecase(querier, shipService)
	a.ConfigUC = config.NewUsecase(querier, a.SqlPool, a.Cfg)

	dbUrl := a.Cfg.DirectURL
	if dbUrl == "" {
		dbUrl = a.Cfg.DatabaseURL
	}
	store, err := whatsapp.NewStore(dbUrl)
	if err != nil {
		return fmt.Errorf("whatsapp store init: %w", err)
	}
	a.WAStore = store

	receipt.InitProcessor()

	a.BotManager = whatsapp.NewManager(a.Context, a.Cfg, a.ShipmentUC, a.ConfigUC, a.WAStore, &a.WG, cache.AsynqClient)

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

	notif.InitMailer(a.Cfg)

	a.HttpServer = transport_http.NewServer(a.Cfg, a.ShipmentUC, a.ConfigUC, a.SqlPool, a)

	// Parse Redis connection options for Asynq Server/Scheduler (client reuses cache.AsynqClient)
	opt, err := asynq.ParseRedisURI(a.Cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("asynq redis URI parse: %w", err)
	}
	redisOpt, ok := opt.(asynq.RedisClientOpt)
	if !ok {
		return fmt.Errorf("unexpected asynq redis option type: %T", opt)
	}
	redisConnOpt := asynq.RedisClientOpt{Addr: redisOpt.Addr, Password: redisOpt.Password, DB: redisOpt.DB}

	// Initialize the global Worker — reuse the shared Asynq client to avoid connection leak
	a.Worker = &worker.Worker{
		ID:              0, // Global worker
		Bots:            a.BotManager,
		ShipmentUC:      a.ShipmentUC,
		ConfigUC:        a.ConfigUC,
		Cfg:             a.Cfg,
		FrontendURL:     a.Cfg.FrontendURL,
		ShipmentService: a.ShipmentUC.GetService(),
		Context:         a.Context,
		AsynqClient:     cache.AsynqClient,
		Queries:         querier,
	}

	poolSize := a.Cfg.WorkerPoolSize
	if poolSize <= 0 {
		poolSize = 150 // Scaled default up for enterprise I/O workloads
	}

	a.AsynqServer = asynq.NewServer(
		redisConnOpt,
		asynq.Config{
			Concurrency: poolSize,
			Queues: map[string]int{
				"default": 10,
			},
			Logger: logger.AsynqLogger{},
		},
	)

	a.AsynqScheduler = asynq.NewScheduler(
		redisConnOpt,
		&asynq.SchedulerOpts{
			Logger: logger.AsynqLogger{},
		},
	)

	return nil
}

func (a *App) Run() error {
	// Register cron tasks
	a.AsynqScheduler.Register("*/5 * * * *", asynq.NewTask(tasks.TypeCronPulse, nil))
	a.AsynqScheduler.Register("0 8 * * *", asynq.NewTask(tasks.TypeCronDailyStats, nil))
	a.AsynqScheduler.Register("0 0 * * *", asynq.NewTask(tasks.TypeCronPruning, nil))
	a.AsynqScheduler.Register("*/10 * * * *", asynq.NewTask(tasks.TypeCronHealthCheck, nil))
	a.AsynqScheduler.Register("*/3 * * * *", asynq.NewTask(tasks.TypeCronBotLiveness, nil))

	go func() {
		if err := a.AsynqScheduler.Start(); err != nil {
			logger.Error().Err(err).Msg("Asynq Scheduler failed to start")
		}
	}()

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

	// Fix Thundering Herd: Connect bots concurrently in batches to drastically improve server startup time for thousands of tenants.
	go func() {
		bots := a.BotManager.GetAllBots()
		batchSize := 50
		for i := 0; i < len(bots); i += batchSize {
			end := i + batchSize
			if end > len(bots) {
				end = len(bots)
			}
			batch := bots[i:end]
			var batchWg sync.WaitGroup

			for _, bot := range batch {
				batchWg.Add(1)
				go func(b models.BotInstance) {
					defer batchWg.Done()
					wc := b.GetWAClient()
					if wc != nil && wc.Store != nil && wc.Store.ID != nil {
						if err := wc.Connect(); err != nil {
							logger.Error().Err(err).Str("company", b.GetCompanyName()).Msg("Failed to connect bot on startup")
						}
					}
				}(bot)
			}
			// Wait for current batch to finish connecting before starting next batch
			batchWg.Wait()
			// Tiny jitter between batches to prevent overwhelming WhatsApp edge servers
			time.Sleep(500 * time.Millisecond)
		}
	}()

	go func() {
		if err := a.HttpServer.Start(a.Cfg.APIPort); err != nil {
			logger.Error().Err(err).Msg("HTTP Server startup failed")
		}
	}()

	go func() {
		mux := asynq.NewServeMux()
		mux.HandleFunc(tasks.TypeWhatsAppMessage, a.Worker.HandleWhatsAppMessage)
		mux.HandleFunc(tasks.TypeCronPulse, a.Worker.HandleCronPulse)
		mux.HandleFunc(tasks.TypeCronDailyStats, a.Worker.HandleCronDailyStats)
		mux.HandleFunc(tasks.TypeCronPruning, a.Worker.HandleCronPruning)
		mux.HandleFunc(tasks.TypeCronHealthCheck, a.Worker.HandleCronHealthCheck)
		mux.HandleFunc(tasks.TypeCronBotLiveness, a.Worker.HandleCronBotLiveness)
		mux.HandleFunc(tasks.TypeOutboundAlert, a.Worker.HandleOutboundAlert)
		mux.HandleFunc(tasks.TypeCronCompanyPulse, a.Worker.HandleCronCompanyPulse)

		if err := a.AsynqServer.Run(mux); err != nil {
			logger.Error().Err(err).Msg("Asynq Server failed")
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
	}

	if a.AsynqScheduler != nil {
		a.AsynqScheduler.Shutdown()
	}

	if a.AsynqServer != nil {
		a.AsynqServer.Stop()
	}

	receipt.Shutdown()
	notif.ShutdownMailer()

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

func (a *App) LivenessCheck() {
	a.BotManager.LivenessCheck()
}

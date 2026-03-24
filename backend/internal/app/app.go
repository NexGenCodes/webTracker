package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"webtracker-bot/internal/commands"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/scheduler"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/utils"
	"webtracker-bot/internal/whatsapp"
	"webtracker-bot/internal/worker"
	"database/sql"

	"webtracker-bot/internal/adapter/db"
	"webtracker-bot/internal/adapter/dbconn"
	transport_http "webtracker-bot/internal/transport/http"
	"webtracker-bot/internal/usecase"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type App struct {
	Cfg        *config.Config
	ShipmentUC *usecase.ShipmentUsecase
	ConfigUC   *usecase.ConfigUsecase
	WA         *whatsmeow.Client
	Jobs       chan models.Job
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

	wa, err := whatsapp.NewClient(a.Cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("whatsapp init (Postgres): %w", err)
	}
	a.WA = wa

	worker.InitReceiptProcessor(a.Cfg.CompanyName, a.ShipmentUC, whatsapp.NewSender(a.WA, a.Cfg.CompanyName))

	a.WA.AddEventHandler(a.handleWAEvent)

	if err := utils.InitReceiptRenderer(); err != nil {
		logger.Error().Err(err).Msg("Failed to initialize receipt renderer (Font download failed?). Receipts may look generic.")
	} else {
		logger.Info().Msg("Receipt renderer initialized (Fonts loaded)")
	}

	// Init Fiber HTTP REST API Server
	a.HttpServer = transport_http.NewServer(a.Cfg, a.ShipmentUC, a.SqlPool)

	return nil
}

func (a *App) Run() error {
	sender := whatsapp.NewSender(a.WA, a.Cfg.CompanyName)
	cmdDispatcher := commands.NewDispatcher(a.ShipmentUC, a.ConfigUC, sender, a.Cfg.CompanyPrefix, a.Cfg.CompanyName, a.Cfg.PairingPhone, a.Cfg.AdminTimezone)

	shipmentService := a.ShipmentUC.Service

	for i := 1; i <= 10; i++ {
		a.WG.Add(1)
		w := &worker.Worker{
			ID:              i,
			Client:          a.WA,
			Sender:          sender,
			Jobs:            a.Jobs,
			WG:              &a.WG,
			Cfg:             a.Cfg,
			AwbCmd:          a.Cfg.CompanyPrefix,
			CompanyName:     a.Cfg.CompanyName,
			Cmd:             cmdDispatcher,
			ShipmentUC:      a.ShipmentUC,
			ConfigUC:        a.ConfigUC,
			TrackingBaseURL: a.Cfg.TrackingBaseURL,
			ShipmentService: shipmentService,
		}
		go w.Start()
	}


	a.Cron = scheduler.NewManager(a.Cfg, a.ShipmentUC, a.ConfigUC, a.WA)
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
	if a.WA.Store.ID == nil {
		if a.Cfg.PairingPhone == "" {
			return fmt.Errorf("not logged in and no WHATSAPP_PAIRING_PHONE provided in .env")
		}

		if err := a.WA.Connect(); err != nil {
			return err
		}

		// Wait for connection to stabilize
		time.Sleep(5 * time.Second)

		code, err := a.WA.PairPhone(a.Context, a.Cfg.PairingPhone, true, whatsmeow.PairClientChrome, "Chrome (Windows)")
		if err != nil {
			return err
		}

		logger.Info().Str("code", code).Msg("Pairing code generated")
		fmt.Println("********************************")
		fmt.Println("*                              *")
		fmt.Println("*   PAIRING CODE: " + code + "   *")
		fmt.Println("*                              *")
		fmt.Println("********************************")

		// Send via Email (if configured)
		notif.SendPairingCodeEmail(a.Cfg, code)
	} else {
		return a.WA.Connect()
	}
	return nil
}

func (a *App) handleWAEvent(evt interface{}) {
	whatsapp.HandleEvent(a.WA, evt, a.Jobs, a.Cfg, a.ConfigUC)

	switch evt.(type) {
	case *events.Connected:
		logger.Info().Msg("WhatsApp Connected!")
	case *events.LoggedOut:
		logger.Warn().Msg("WhatsApp Logged Out!")
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
	if a.WA != nil {
		a.WA.Disconnect()
		logger.Info().Msg("WhatsApp disconnected")
	}

	// 3. Stop Scheduler
	if a.Cron != nil {
		a.Cron.Stop()
		logger.Info().Msg("Cron scheduler stopped")
	}

	// 4. Close Worker Channel and Wait
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

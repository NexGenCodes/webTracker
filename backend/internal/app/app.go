package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"webtracker-bot/internal/commands"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/health"
	"webtracker-bot/internal/localdb"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/scheduler"
	"webtracker-bot/internal/supabase"
	"webtracker-bot/internal/utils"
	"webtracker-bot/internal/whatsapp"
	"webtracker-bot/internal/worker"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type App struct {
	Cfg     *config.Config
	DB      *supabase.Client
	LocalDB *localdb.Client
	WA      *whatsmeow.Client
	Jobs    chan models.Job
	WG      sync.WaitGroup
	Cancel  context.CancelFunc
	Cron    *scheduler.CronManager
	Context context.Context
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
	// 0. Verify Environment
	if err := health.VerifyEnvironment(); err != nil {
		fmt.Printf("FATAL: Environment verification failed: %v\n", err)
		os.Exit(1)
	}

	// 1. Init DB
	db, err := supabase.NewClient(a.Cfg.DatabaseURL, a.Cfg.CompanyPrefix)
	if err != nil {
		return fmt.Errorf("db init: %w", err)
	}
	a.DB = db

	// 1.5 Init LocalDB (SQLite)
	ldb, err := localdb.NewClient(a.Cfg.WhatsAppSessionPath)
	if err != nil {
		return fmt.Errorf("localdb init: %w", err)
	}
	a.LocalDB = ldb

	// 2. Init WhatsApp
	wa, err := whatsapp.NewClient(a.Cfg.WhatsAppSessionPath)
	if err != nil {
		return fmt.Errorf("whatsapp init: %w", err)
	}
	a.WA = wa

	// 3. Register Events
	a.WA.AddEventHandler(a.handleWAEvent)

	// 4. Init Receipt Renderer (Native)
	if err := utils.InitReceiptRenderer(); err != nil {
		logger.Error().Err(err).Msg("Failed to initialize receipt renderer (Font download failed?). Receipts may look generic.")
	} else {
		logger.Info().Msg("Receipt renderer initialized (Fonts loaded)")
	}

	// 5. Start Health Server
	health.StartHealthServer("8080", func() error {
		return a.DB.Ping()
	})

	return nil
}

func (a *App) Run() error {
	// Start Workers
	cmdDispatcher := commands.NewDispatcher(a.DB, a.LocalDB, a.Cfg.CompanyPrefix, a.Cfg.CompanyName)
	sender := whatsapp.NewSender(a.WA)
	for i := 1; i <= 5; i++ {
		a.WG.Add(1)
		w := &worker.Worker{
			ID:          i,
			Client:      a.WA,
			Sender:      sender,
			DB:          a.DB,
			Jobs:        a.Jobs,
			WG:          &a.WG,
			GeminiKey:   a.Cfg.GeminiAPIKey,
			AwbCmd:      a.Cfg.CompanyPrefix,
			CompanyName: a.Cfg.CompanyName,
			Cmd:         cmdDispatcher,
			LocalDB:     a.LocalDB,
		}
		go w.Start()
	}

	// Start Scheduler
	scheduler.StartDailySummary(a.WA, a.DB, a.Cfg.AdminTimezone, a.Cfg.AllowedGroups)
	a.Cron = scheduler.NewManager(a.Cfg, a.DB, a.WA)
	a.Cron.Start()

	// Connect to WhatsApp
	if err := a.connectWA(); err != nil {
		return err
	}

	// Wait for Signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		logger.Info().Str("signal", sig.String()).Msg("Shutdown signal received")
	case <-a.Context.Done():
		logger.Warn().Msg("App context cancelled")
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

		code, err := a.WA.PairPhone(a.Context, a.Cfg.PairingPhone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
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
	whatsapp.HandleEvent(a.WA, evt, a.Jobs, a.Cfg, a.DB)

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
	a.Cancel()
	a.WA.Disconnect()
	if a.Cron != nil {
		a.Cron.Stop()
	}
	close(a.Jobs)
	a.WG.Wait()
	logger.Info().Msg("Bot shutdown complete.")
	return nil
}

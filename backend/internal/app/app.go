package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"webtracker-bot/internal/api"
	"webtracker-bot/internal/commands"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/localdb"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/scheduler"
	"webtracker-bot/internal/utils"
	"webtracker-bot/internal/whatsapp"
	"webtracker-bot/internal/worker"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type App struct {
	Cfg     *config.Config
	LocalDB *localdb.Client
	WA      *whatsmeow.Client
	Jobs    chan models.Job
	WG      sync.WaitGroup
	Cancel  context.CancelFunc
	Cron    *scheduler.CronManager
	API     *api.Server
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
	workDir := config.GetWorkDir()
	dbPath := filepath.Join(workDir, "webtracker.db")
	ldb, err := localdb.NewClient(dbPath)
	if err != nil {
		return fmt.Errorf("localdb init: %w", err)
	}
	a.LocalDB = ldb

	wa, err := whatsapp.NewClient(a.Cfg.WhatsAppSessionPath)
	if err != nil {
		return fmt.Errorf("whatsapp init: %w", err)
	}
	a.WA = wa

	a.WA.AddEventHandler(a.handleWAEvent)

	if err := utils.InitReceiptRenderer(); err != nil {
		logger.Error().Err(err).Msg("Failed to initialize receipt renderer (Font download failed?). Receipts may look generic.")
	} else {
		logger.Info().Msg("Receipt renderer initialized (Fonts loaded)")
	}

	// 5. API Server initialized during Run()

	return nil
}

func (a *App) Run() error {
	// Start Workers
	sender := whatsapp.NewSender(a.WA)
	cmdDispatcher := commands.NewDispatcher(a.LocalDB, sender, a.Cfg.CompanyPrefix, a.Cfg.CompanyName, a.Cfg.PairingPhone, a.Cfg.AdminTimezone)
	for i := 1; i <= 5; i++ {
		a.WG.Add(1)
		w := &worker.Worker{
			ID:              i,
			Client:          a.WA,
			Sender:          sender,
			Jobs:            a.Jobs,
			WG:              &a.WG,
			GeminiKey:       a.Cfg.GeminiAPIKey,
			AwbCmd:          a.Cfg.CompanyPrefix,
			CompanyName:     a.Cfg.CompanyName,
			Cmd:             cmdDispatcher,
			LocalDB:         a.LocalDB,
			TrackingBaseURL: a.Cfg.TrackingBaseURL,
		}
		go w.Start()
	}

	port := a.Cfg.APIPort
	if port == "" {
		port = "8080"
	}

	a.API = api.NewServer(a.LocalDB, a.Cfg.ApiAuthToken, a.Cfg.GeminiAPIKey, a.Cfg.AllowedOrigin)
	go func() {
		logger.Info().Str("port", port).Msg("API Server starting")
		if err := a.API.Start(port); err != nil {
			logger.Error().Err(err).Msg("API Server failed")
		}
	}()

	a.Cron = scheduler.NewManager(a.Cfg, a.LocalDB, a.WA)
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
	whatsapp.HandleEvent(a.WA, evt, a.Jobs, a.Cfg, a.LocalDB)

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

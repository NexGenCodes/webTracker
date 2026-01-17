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
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/scheduler"
	"webtracker-bot/internal/supabase"
	"webtracker-bot/internal/whatsapp"
	"webtracker-bot/internal/worker"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type App struct {
	Cfg     *config.Config
	DB      *supabase.Client
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
	// 1. Init DB
	db, err := supabase.NewClient(a.Cfg.DatabaseURL, a.Cfg.CompanyPrefix)
	if err != nil {
		return fmt.Errorf("db init: %w", err)
	}
	a.DB = db

	// 2. Init WhatsApp
	wa, err := whatsapp.NewClient(a.Cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("whatsapp init: %w", err)
	}
	a.WA = wa

	// 3. Register Events
	a.WA.AddEventHandler(a.handleWAEvent)

	return nil
}

func (a *App) Run() error {
	// Start Workers
	cmdDispatcher := commands.NewDispatcher(a.DB, a.Cfg.CompanyPrefix, a.Cfg.CompanyName)
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
		qrChan, _ := a.WA.GetQRChannel(a.Context)
		if err := a.WA.Connect(); err != nil {
			return err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("\nScan the QR code above to login")
			}
		}
	} else {
		return a.WA.Connect()
	}
	return nil
}

func (a *App) handleWAEvent(evt interface{}) {
	whatsapp.HandleEvent(evt, a.Jobs, a.Cfg.AllowedGroups)

	switch evt.(type) {
	case *events.QR:
		logger.Info().Msg("QR Code received")
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

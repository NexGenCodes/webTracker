package main

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
	"go.mau.fi/whatsmeow/types/events"
)

func main() {
	// Load Config
	cfg := config.Load()

	// Init Logger
	logger.Init()

	// Validate Config (Pre-flight)
	if err := cfg.Validate(); err != nil {
		logger.Fatal().Err(err).Msg("Configuration validation failed")
	}

	// Create Root Context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 5. Init DB
	dbClient, err := supabase.NewClient(cfg.DatabaseURL, cfg.CompanyPrefix)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to init Database")
	}

	// 4. Init WhatsApp
	client, err := whatsapp.NewClient(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to init WhatsApp")
	}

	// 5. Create Command Dispatcher
	cmdDispatcher := commands.NewDispatcher(dbClient, cfg.CompanyPrefix, cfg.CompanyName)

	// 6. Create Job Queue
	jobQueue := make(chan models.Job, cfg.BufferSize)
	var wg sync.WaitGroup

	// Start Workers
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		w := &worker.Worker{
			ID:          i,
			Client:      client,
			DB:          dbClient,
			Jobs:        jobQueue,
			WG:          &wg,
			GeminiKey:   cfg.GeminiAPIKey,
			AwbCmd:      cfg.CompanyPrefix,
			CompanyName: cfg.CompanyName,
			Cmd:         cmdDispatcher,
		}
		go w.Start()
	}

	// 7. Register Event Handler
	client.AddEventHandler(func(evt interface{}) {
		whatsapp.HandleEvent(evt, jobQueue, cfg.AllowedGroups)

		// Handle QR codes and connection events
		switch evt.(type) {
		case *events.QR:
			logger.Info().Msg("QR Code received. Please scan.")
		case *events.Connected:
			logger.Info().Msg("WhatsApp Connected!")
		case *events.LoggedOut:
			logger.Warn().Msg("WhatsApp Logged Out! Please re-scan.")
			cancel() // Signal shutdown on logout
		}
	})

	// 8. Start Scheduler (Native Maintenance + Stats)
	scheduler.StartDailySummary(client, dbClient, cfg.AdminTimezone, cfg.AllowedGroups)
	cronMgr := scheduler.NewManager(cfg, dbClient, client)
	cronMgr.Start()

	// 9. Connect
	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to connect")
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("Scan the QR code above to login")
			} else {
				fmt.Println("QR Event:", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to reconnect")
		}
	}

	// Shutdown Signal Handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	logger.Info().Msg("Bot is running. Press CTRL+C to exit.")

	select {
	case sig := <-sigChan:
		logger.Info().Str("signal", sig.String()).Msg("Received termination signal")
	case <-ctx.Done():
		logger.Warn().Msg("Root context cancelled (Logged out or internal failure)")
	}

	logger.Info().Msg("Graceful shutdown initiated...")
	cancel() // Cancel context to stop workers/scheduler

	client.Disconnect()
	cronMgr.Stop()
	close(jobQueue)
	wg.Wait()
	logger.Info().Msg("Bot shutdown complete.")
}

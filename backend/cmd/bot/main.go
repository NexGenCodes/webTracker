package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/scheduler"
	"webtracker-bot/internal/supabase"
	"webtracker-bot/internal/whatsapp"
	"webtracker-bot/internal/worker"

	"go.mau.fi/whatsmeow/types/events"
)

func main() {
	// 1. Load Config (also loads .env into os env)
	cfg := config.Load()

	// 2. Init Logger (reads LOG_PATH from env)
	logger.Init()

	// 3. Init DB
	dbClient, err := supabase.NewClient(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to init Database")
	}

	// 4. Init WhatsApp
	client, err := whatsapp.NewClient(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to init WhatsApp")
	}

	// 5. Create Job Queue
	jobQueue := make(chan models.Job, cfg.BufferSize)
	var wg sync.WaitGroup

	// 6. Start Workers (Scaled to 5 as requested)
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		w := &worker.Worker{
			ID:        i,
			Client:    client,
			DB:        dbClient,
			Jobs:      jobQueue,
			WG:        &wg,
			GeminiKey: cfg.GeminiAPIKey,
		}
		go w.Start()
	}

	// 7. Register Event Handler
	client.AddEventHandler(func(evt interface{}) {
		whatsapp.HandleEvent(evt, jobQueue, cfg.GroupID)

		// Handle QR codes and connection events
		switch evt.(type) {
		case *events.QR:
			logger.Info().Msg("QR Code received. Please scan.")
		case *events.Connected:
			logger.Info().Msg("WhatsApp Connected!")
		}
	})

	// 8. Start Scheduler (Native Maintenance + Stats)
	scheduler.StartDailySummary(client, dbClient, cfg.AdminTimezone, cfg.GroupID)
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
				fmt.Println("QR Code:", evt.Code)
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

	// 10. Shutdown Signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	logger.Info().Msg("Bot is running. Press CTRL+C to exit.")
	<-c

	logger.Info().Msg("Shutting down...")
	client.Disconnect()
	close(jobQueue)
	wg.Wait()
	logger.Info().Msg("Shutdown complete.")
}

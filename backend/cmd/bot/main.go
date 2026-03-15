package main

import (
	"context"
	"flag"
	"os"
	"runtime/debug"
	"time"
	"webtracker-bot/internal/app"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/localdb"
	"webtracker-bot/internal/logger"
)

func main() {
	resetDB := flag.Bool("reset", false, "Reset the database and exit")
	flag.Parse()
	debug.SetGCPercent(50)
	debug.SetMemoryLimit(700 * 1024 * 1024)
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			debug.FreeOSMemory()
		}
	}()

	// 1. Load Config
	cfg := config.Load()

	// 2. Init Logger
	logger.Init()

	// 3. Validate Config
	if err := cfg.Validate(); err != nil {
		logger.Fatal().Err(err).Msg("Configuration validation failed")
	}

	// 3.1 Handle Reset Flag
	if *resetDB {
		ldb, err := localdb.NewClient(cfg.DatabaseURL)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to connect to DB for reset")
		}
		defer ldb.Close()

		logger.Info().Msg("🚀 Resetting database...")
		if err := ldb.ResetDB(context.Background()); err != nil {
			logger.Fatal().Err(err).Msg("Failed to reset database")
		}
		logger.Info().Msg("✅ Database reset successfully")
		os.Exit(0)
	}

	// 4. Initialize App
	application := app.New(cfg)
	if err := application.Init(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize app")
	}

	// 5. Run App
	logger.Info().Msg("Bot starting...")
	if err := application.Run(); err != nil {
		logger.Fatal().Err(err).Msg("App crashed")
	}
}

package main

import (
	"webtracker-bot/internal/app"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
)

func main() {
	// 1. Load Config
	cfg := config.Load()

	// 2. Init Logger
	logger.Init()

	// 3. Validate Config
	if err := cfg.Validate(); err != nil {
		logger.Fatal().Err(err).Msg("Configuration validation failed")
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

package main

import (
	"flag"
	"os"
	"runtime/debug"
	"strconv"
	"time"
	"webtracker-bot/internal/app"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
)

func main() {
	flag.Parse()

	// Dynamic Vertical Scaling Configuration
	// By default, it limits to 700MB. To scale vertically, set MAX_MEMORY_MB in your environment.
	// e.g. MAX_MEMORY_MB=2048 for a 2GB VPS.
	memLimitMB := int64(700)
	if envMem := os.Getenv("MAX_MEMORY_MB"); envMem != "" {
		if parsed, err := strconv.ParseInt(envMem, 10, 64); err == nil {
			memLimitMB = parsed
		}
	}

	debug.SetGCPercent(50)
	debug.SetMemoryLimit(memLimitMB * 1024 * 1024)

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

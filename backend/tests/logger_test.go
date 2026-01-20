package tests

import (
	"os"
	"path/filepath"
	"testing"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
)

func TestLoggerInit(t *testing.T) {
	// Use real config
	cfg, _ := config.LoadFromEnv()
	logPath := "./test_logs"
	if cfg != nil && cfg.LogPath != "" {
		logPath = cfg.LogPath
	}

	// Ensure we don't delete real logs if they are important
	// but the test environment should be isolated.
	// For now, we follow the "use main config" instruction.

	logger.Init()

	logger.Info().Msg("Test info message")
	logger.Error().Msg("Test error message")

	// Check if log file was created
	if _, err := os.Stat(filepath.Join(logPath, "bot.log")); os.IsNotExist(err) {
		t.Errorf("Expected log file bot.log to be created in %s", logPath)
	}
}

func TestLoggerLevels(t *testing.T) {
	// Verify logger doesn't panic on different levels
	logger.Debug().Msg("Debug")
	logger.Warn().Msg("Warn")
}

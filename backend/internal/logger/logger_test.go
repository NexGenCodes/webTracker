package logger

import (
	"os"
	"testing"
)

func TestLoggerInit(t *testing.T) {
	// Test with log path
	os.Setenv("LOG_PATH", "test_logs")
	defer os.RemoveAll("test_logs")
	defer os.Unsetenv("LOG_PATH")

	Init()

	Info().Msg("Test info message")
	l := WithField("module", "test")
	l.Info().Msg("Test contextual log")
	l2 := WithFields(map[string]interface{}{"id": 123, "user": "test"})
	l2.Warn().Msg("Test multi-field log")
}

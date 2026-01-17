package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Init() {
	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339

	var writers []io.Writer

	// 1. Console Writer
	if os.Getenv("GO_ENV") != "production" {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	} else {
		writers = append(writers, os.Stderr)
	}

	// 2. File Writer (Log Rotation)
	logPath := os.Getenv("LOG_PATH")
	if logPath != "" {
		// Ensure directory exists
		_ = os.MkdirAll(logPath, 0744)

		fileLogger := &lumberjack.Logger{
			Filename:   filepath.Join(logPath, "bot.log"),
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   // days
			Compress:   true, // disabled by default
		}
		writers = append(writers, fileLogger)
	}

	multi := zerolog.MultiLevelWriter(writers...)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()

	// Set global log level
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	l, err := zerolog.ParseLevel(level)
	if err != nil {
		l = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(l)
}

func Info() *zerolog.Event {
	return log.Info()
}

func Error() *zerolog.Event {
	return log.Error()
}

func Warn() *zerolog.Event {
	return log.Warn()
}

func Debug() *zerolog.Event {
	return log.Debug()
}

func Fatal() *zerolog.Event {
	return log.Fatal()
}

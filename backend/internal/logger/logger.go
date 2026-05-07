package logger

import (
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ServiceName is stamped onto every log line.
// Override before calling Init() if you run multiple binaries from the same module.
var ServiceName = "webtracker-bot"

// BuildVersion is injected at link-time via -ldflags.
// e.g. go build -ldflags "-X webtracker-bot/internal/logger.BuildVersion=$(git rev-parse --short HEAD)"
var BuildVersion = "dev"

// Init initialises the global zerolog logger.
// Must be called once at application startup, before any logging.
func Init() {
	// --- Time format ---
	zerolog.TimeFieldFormat = time.RFC3339

	// --- Callers: trim to module-relative paths so logs are shorter ---
	zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
		// Keep only the last two path segments: "package/file.go:line"
		short := filepath.Base(filepath.Dir(file)) + "/" + filepath.Base(file)
		return short + ":" + itoa(line)
	}

	// --- Writers ---
	var writers []io.Writer

	// 1. Console Writer (human-readable in dev, raw JSON in prod)
	if os.Getenv("GO_ENV") != "production" {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
		})
	} else {
		// Raw JSON to stderr → captured by systemd/journald/Docker log driver
		writers = append(writers, os.Stderr)
	}

	// 2. Rotating file writer
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "logs"
	}
	_ = os.MkdirAll(logPath, 0o744)
	fileLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logPath, "bot.log"),
		MaxSize:    10, // MB per file
		MaxBackups: 5,  // keep 5 rotated files
		MaxAge:     28, // days
		Compress:   true,
	}
	writers = append(writers, fileLogger)

	// --- Build global context ---
	hostname, _ := os.Hostname()

	// Detect build version from embedded VCS info when not set via -ldflags
	if BuildVersion == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, s := range info.Settings {
				if s.Key == "vcs.revision" && len(s.Value) >= 7 {
					BuildVersion = s.Value[:7]
				}
			}
		}
	}

	multi := zerolog.MultiLevelWriter(writers...)
	log.Logger = zerolog.New(multi).With().
		Timestamp().
		Caller().
		Str("service", ServiceName).
		Str("host", hostname).
		Str("version", BuildVersion).
		Logger()

	// --- Log Level ---
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

// itoa is a zero-alloc int → string converter used by CallerMarshalFunc.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

// --- Level helpers (thin wrappers so callers never import zerolog directly) ---

func Trace() *zerolog.Event { return log.Trace() }
func Debug() *zerolog.Event { return log.Debug() }
func Info() *zerolog.Event  { return log.Info() }
func Warn() *zerolog.Event  { return log.Warn() }
func Error() *zerolog.Event { return log.Error() }
func Fatal() *zerolog.Event { return log.Fatal() }

// Panic logs at Panic level and then panics.
func Panic() *zerolog.Event { return log.Panic() }

// --- Context helpers ---

// WithComponent returns a zerolog.Logger pre-stamped with a "component" field.
// Use it to create a package-level logger that carries its own label:
//
//	var log = logger.WithComponent("whatsapp.sender")
func WithComponent(name string) zerolog.Logger {
	return log.Logger.With().Str("component", name).Logger()
}

// WithField adds a single field to the log
func WithField(key string, value interface{}) zerolog.Logger {
	return log.With().Interface(key, value).Logger()
}

// WithFields adds multiple fields to the log
func WithFields(fields map[string]interface{}) zerolog.Logger {
	return log.With().Fields(fields).Logger()
}

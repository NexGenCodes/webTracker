package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	GroupID            string
	GeminiAPIKey       string
	AdminTimezone      string
	AppURL             string
	ExternalCronSecret string
	HealthcheckURL     string
	LogPath            string
	WorkerPoolSize     int
	BufferSize         int
}

func GetWorkDir() string {
	ex, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(ex)
}

func Load() *Config {
	workDir := GetWorkDir()

	// Try loading from executable dir
	_ = godotenv.Load(filepath.Join(workDir, ".env.local"))
	_ = godotenv.Load(filepath.Join(workDir, ".env"))

	// Fallback to current dir for local dev
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:        os.Getenv("DIRECT_URL"),
		GroupID:            os.Getenv("WHATSAPP_GROUP_ID"),
		GeminiAPIKey:       os.Getenv("GEMINI_API_KEY"),
		AdminTimezone:      os.Getenv("ADMIN_TIMEZONE"),
		AppURL:             os.Getenv("APP_URL"),
		ExternalCronSecret: os.Getenv("EXTERNAL_CRON_SECRET"),
		HealthcheckURL:     os.Getenv("HEALTHCHECK_URL"),
		LogPath:            os.Getenv("LOG_PATH"),
		WorkerPoolSize:     5,
		BufferSize:         100,
	}

	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	}

	if cfg.LogPath == "" {
		cfg.LogPath = filepath.Join(workDir, "logs")
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("CRITICAL: DATABASE_URL or DIRECT_URL must be set.")
	}

	return cfg
}

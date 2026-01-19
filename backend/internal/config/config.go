package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	AllowedGroups  []string
	CompanyPrefix  string
	GeminiAPIKey   string
	AdminTimezone  string
	HealthcheckURL string
	LogPath        string
	CompanyName    string
	WorkerPoolSize int
	BufferSize     int
}

func GetWorkDir() string {
	ex, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(ex)
}

func generateAbbreviation(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "AWB"
	}

	parts := strings.Fields(name)
	if len(parts) > 1 {
		abbr := ""
		for _, p := range parts {
			if len(p) > 0 {
				abbr += string(p[0])
			}
		}
		return strings.ToUpper(abbr)
	}

	if len(name) > 3 {
		return strings.ToUpper(name[:3])
	}
	return strings.ToUpper(name)
}

func Load() *Config {
	workDir := GetWorkDir()

	// Try loading from executable dir
	_ = godotenv.Load(filepath.Join(workDir, ".env.local"))
	_ = godotenv.Load(filepath.Join(workDir, ".env"))

	// Fallback to current dir for local dev
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load()

	allowedStr := os.Getenv("WHATSAPP_ALLOWED_GROUPS")
	var allowedGroups []string
	if allowedStr != "" {
		allowedGroups = strings.Split(allowedStr, ",")
		for i, s := range allowedGroups {
			allowedGroups[i] = strings.TrimSpace(s)
		}
	}

	companyNameEnv := os.Getenv("COMPANY_NAME")
	companyName := strings.ReplaceAll(companyNameEnv, " ", "")
	if companyName == "" {
		companyName = "Airwaybill"
	}
	companyPrefix := generateAbbreviation(companyName)

	cfg := &Config{
		DatabaseURL:    os.Getenv("DIRECT_URL"),
		AllowedGroups:  allowedGroups,
		CompanyPrefix:  companyPrefix,
		GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
		AdminTimezone:  os.Getenv("ADMIN_TIMEZONE"),
		HealthcheckURL: os.Getenv("HEALTHCHECK_URL"),
		LogPath:        os.Getenv("LOG_PATH"),
		CompanyName:    companyName,
		WorkerPoolSize: 5,
		BufferSize:     100,
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

func (cfg *Config) Validate() error {
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DIRECT_URL or DATABASE_URL is missing")
	}
	if cfg.GeminiAPIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY is missing")
	}
	if cfg.CompanyPrefix == "" {
		cfg.CompanyPrefix = "AWB"
	}
	return nil
}

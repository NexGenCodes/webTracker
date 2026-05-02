package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string `env:"DATABASE_URL"`
	DirectURL      string `env:"DIRECT_URL"`
	GeminiAPIKey   string `env:"GEMINI_API_KEY"`
	AdminTimezone  string `env:"ADMIN_TIMEZONE" env-default:"Africa/Lagos"`
	HealthcheckURL string `env:"HEALTHCHECK_URL"`
	LogPath        string `env:"LOG_PATH"`
	LogLevel       string `env:"LOG_LEVEL" env-default:"info"`
	WorkerPoolSize int    `env:"WORKER_POOL_SIZE" env-default:"5"`
	BufferSize     int    `env:"BUFFER_SIZE" env-default:"100"`

	// Notification Config
	SMTPHost     string `env:"SMTP_HOST"`
	SMTPPort     int    `env:"SMTP_PORT" env-default:"587"`
	SMTPUsername string `env:"SMTP_USERNAME"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	NotifyEmail  string `env:"NOTIFY_EMAIL"`

	// Receipt Configuration
	UseOptimizedReceipt bool `env:"USE_OPTIMIZED_RECEIPT" env-default:"true"`

	// Access Control
	AllowPrivateChat bool `env:"WHATSAPP_ALLOW_PRIVATE_CHAT" env-default:"false"`

	// REST API Port
	APIPort string `env:"API_PORT" env-default:"5000"`

	// Paystack
	PaystackSecretKey string `env:"PAYSTACK_SECRET_KEY"`

	// JWT Authentication
	JWTSecret         string `env:"JWT_SECRET"`
	JWTPrivateKeyPath string `env:"JWT_PRIVATE_KEY_PATH" env-default:"jwt_private.pem"`
	JWTPublicKeyPath  string `env:"JWT_PUBLIC_KEY_PATH" env-default:"jwt_public.pem"`

	// Frontend URL for magic links
	FrontendURL string `env:"FRONTEND_URL" env-default:"http://localhost:3000"`

	// Super Admin — bypasses all billing, unlimited shipments
	SuperAdminCompanyID string `env:"SUPERADMIN_COMPANY_ID"`
}

func GetWorkDir() string {
	dir, _ := os.Getwd()
	for dir != "" && dir != "." && dir != "/" {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}

	ex, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(ex)
}

func Load() *Config {
	workDir := GetWorkDir()

	// Load .env files manually to support multiple paths for different entry points (cmd/bot, cmd/migrate, etc.)
	_ = godotenv.Load(filepath.Join(workDir, ".env.local"))
	_ = godotenv.Load(filepath.Join(workDir, ".env"))
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load()

	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(fmt.Sprintf("Failed to read configuration: %v", err))
	}

	// Post-processing
	// Database Selection (DirectURL preferred for migrations/SQLC types)
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = cfg.DirectURL
	} else if cfg.DirectURL != "" {
		// If both are provided, we usually want DatabaseURL for runtime (pooled)
		// but Config.DatabaseURL in our code is what the app uses.
	}

	// Ensure search_path is set for Neon
	if cfg.DatabaseURL != "" && !strings.Contains(cfg.DatabaseURL, "search_path=public") {
		if strings.Contains(cfg.DatabaseURL, "?") {
			cfg.DatabaseURL += "&search_path=public"
		} else {
			cfg.DatabaseURL += "?search_path=public"
		}
	}

	if cfg.LogPath == "" {
		cfg.LogPath = filepath.Join(workDir, "logs")
	}

	return &cfg
}

func (cfg *Config) Validate() error {
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL / DIRECT_URL is missing")
	}
	if cfg.GeminiAPIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY is missing")
	}
	return nil
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"webtracker-bot/internal/utils"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string `env:"DATABASE_URL"`
	DirectURL      string `env:"DIRECT_URL"`
	CompanyPrefix  string `env:"COMPANY_PREFIX"`
	GeminiAPIKey   string `env:"GEMINI_API_KEY"`
	AdminTimezone  string `env:"ADMIN_TIMEZONE" env-default:"Africa/Lagos"`
	HealthcheckURL string `env:"HEALTHCHECK_URL"`
	LogPath        string `env:"LOG_PATH"`
	LogLevel       string `env:"LOG_LEVEL" env-default:"info"`
	CompanyName    string `env:"COMPANY_NAME" env-default:"Airwaybill"`
	WorkerPoolSize int    `env:"WORKER_POOL_SIZE" env-default:"5"`
	BufferSize     int    `env:"BUFFER_SIZE" env-default:"100"`
	PairingPhone   string `env:"WHATSAPP_PAIRING_PHONE"`
	BotOwnerPhone  string `env:"BOT_OWNER_PHONE"`
	
	// Notification Config
	SMTPHost     string `env:"SMTP_HOST"`
	SMTPPort     int    `env:"SMTP_PORT" env-default:"587"`
	SMTPUsername string `env:"SMTP_USERNAME"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	NotifyEmail  string `env:"NOTIFY_EMAIL"`

	// Access Control
	AllowPrivateChat bool     `env:"WHATSAPP_ALLOW_PRIVATE_CHAT" env-default:"false"`
	AdminPhones      []string `env:"WHATSAPP_ADMIN_PHONES" env-separator:","`

	// Public Tracking URL
	TrackingBaseURL string `env:"TRACKING_BASE_URL" env-default:"http://localhost:3000"`
	
	// REST API Port
	APIPort string `env:"API_PORT" env-default:"5000"`
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

func GenerateAbbreviation(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "AWB"
	}

	reg := regexp.MustCompile("[^a-zA-Z]")
	clean := reg.ReplaceAllString(name, "")
	if clean == "" {
		return "AWB"
	}

	syllables := utils.SplitIntoSyllables(clean)
	count := len(syllables)

	var abbr string
	switch {
	case count == 1:
		s := syllables[0]
		if len(s) <= 3 {
			abbr = s
		} else {
			abbr = string(s[0]) + string(s[len(s)/2]) + string(s[len(s)-1])
		}
	case count == 2:
		s1 := syllables[0]
		s2 := syllables[1]
		p1 := s1
		if len(s1) > 2 {
			p1 = s1[:2]
		}
		p2 := string(s2[0])
		abbr = p1 + p2
	default:
		for i := 0; i < 3 && i < count; i++ {
			abbr += string(syllables[i][0])
		}
	}

	abbr = strings.ToUpper(abbr)
	if len(abbr) > 3 {
		abbr = abbr[:3]
	}
	for len(abbr) < 3 {
		abbr += "X"
	}

	return abbr
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
	cfg.CompanyName = strings.ReplaceAll(cfg.CompanyName, " ", "")
	if cfg.CompanyPrefix == "" {
		cfg.CompanyPrefix = GenerateAbbreviation(cfg.CompanyName)
	}

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

	// Phone Cleaning
	cfg.PairingPhone = cleanPhone(cfg.PairingPhone)
	for i, p := range cfg.AdminPhones {
		cfg.AdminPhones[i] = cleanPhone(p)
	}

	if cfg.BotOwnerPhone == "" && len(cfg.AdminPhones) > 0 {
		cfg.BotOwnerPhone = cfg.AdminPhones[0]
	} else {
		cfg.BotOwnerPhone = cleanPhone(cfg.BotOwnerPhone)
	}

	if cfg.LogPath == "" {
		cfg.LogPath = filepath.Join(workDir, "logs")
	}

	return &cfg
}

func cleanPhone(p string) string {
	p = strings.ReplaceAll(p, "+", "")
	p = strings.ReplaceAll(p, "-", "")
	p = strings.ReplaceAll(p, " ", "")
	return p
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

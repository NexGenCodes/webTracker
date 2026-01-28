package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"webtracker-bot/internal/utils"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	CompanyPrefix  string
	GeminiAPIKey   string
	AdminTimezone  string
	HealthcheckURL string
	LogPath        string
	CompanyName    string
	WorkerPoolSize int
	BufferSize     int
	PairingPhone   string
	BotOwnerPhone  string // New field for notifications
	// Notification Config
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	NotifyEmail  string

	// Access Control
	AllowPrivateChat bool
	AdminPhones      []string

	// Session Storage
	WhatsAppSessionPath string

	// API Auth
	ApiAuthToken string
	APIPort      string

	// Public Tracking URL
	TrackingBaseURL string

	// CORS
	AllowedOrigin string
}

func GetWorkDir() string {
	// Try to find project root by looking for go.mod
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
		// 1 Syllable: First | Middle | Last
		s := syllables[0]
		if len(s) <= 3 {
			abbr = s
		} else {
			abbr = string(s[0]) + string(s[len(s)/2]) + string(s[len(s)-1])
		}
	case count == 2:
		// 2 Syllables: 1st Syllable (2 chars) + 2nd Syllable (1 char)
		s1 := syllables[0]
		s2 := syllables[1]
		p1 := s1
		if len(s1) > 2 {
			p1 = s1[:2]
		}
		p2 := string(s2[0])
		abbr = p1 + p2
	default:
		// 3+ Syllables: 1st char of each of the first 3 syllables
		for i := 0; i < 3 && i < count; i++ {
			abbr += string(syllables[i][0])
		}
	}

	// Pad or truncate to ensure exactly 3 chars if possible
	abbr = strings.ToUpper(abbr)
	if len(abbr) > 3 {
		abbr = abbr[:3]
	}
	for len(abbr) < 3 {
		abbr += "X" // Fallback padding
	}

	return abbr
}

func Load() *Config {
	cfg, err := LoadFromEnv()
	if err != nil {
		log.Fatalf("CRITICAL: %v", err)
	}
	return cfg
}

func LoadFromEnv() (*Config, error) {
	workDir := GetWorkDir()

	// Try loading from executable dir
	_ = godotenv.Load(filepath.Join(workDir, ".env.local"))
	_ = godotenv.Load(filepath.Join(workDir, ".env"))

	// Fallback to current dir for local dev
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load()

	companyNameEnv := os.Getenv("COMPANY_NAME")
	companyName := strings.ReplaceAll(companyNameEnv, " ", "")
	if companyName == "" {
		companyName = "Airwaybill"
	}
	companyPrefix := GenerateAbbreviation(companyName)

	// Parse SMTP Port
	smtpPort := 587
	if p := os.Getenv("SMTP_PORT"); p != "" {
		// simple parsing, ignoring error for brevity in this block
		fmt.Sscanf(p, "%d", &smtpPort)
	}

	var adminPhones []string
	if admins := os.Getenv("WHATSAPP_ADMIN_PHONES"); admins != "" {
		parts := strings.Split(admins, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			p = strings.ReplaceAll(p, "+", "")
			p = strings.ReplaceAll(p, "-", "")
			p = strings.ReplaceAll(p, " ", "")
			if p != "" {
				adminPhones = append(adminPhones, p)
			}
		}
	}

	cfg := &Config{
		DatabaseURL:    os.Getenv("DIRECT_URL"),
		CompanyPrefix:  companyPrefix,
		GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
		AdminTimezone:  os.Getenv("ADMIN_TIMEZONE"),
		HealthcheckURL: os.Getenv("HEALTHCHECK_URL"),
		LogPath:        os.Getenv("LOG_PATH"),
		CompanyName:    companyName,
		WorkerPoolSize: 5,
		BufferSize:     100,
		PairingPhone:   os.Getenv("WHATSAPP_PAIRING_PHONE"),

		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     smtpPort,
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		NotifyEmail:  os.Getenv("NOTIFY_EMAIL"),

		AllowPrivateChat:    os.Getenv("WHATSAPP_ALLOW_PRIVATE_CHAT") == "true",
		AdminPhones:         adminPhones,
		BotOwnerPhone:       os.Getenv("BOT_OWNER_PHONE"), // Load from env
		WhatsAppSessionPath: os.Getenv("WHATSAPP_SESSION_PATH"),
		ApiAuthToken:        os.Getenv("API_AUTH_TOKEN"),
		APIPort:             os.Getenv("API_PORT"),
		TrackingBaseURL:     os.Getenv("TRACKING_BASE_URL"),
		AllowedOrigin:       os.Getenv("ALLOWED_ORIGIN"),
	}

	// Default BotOwnerPhone to first admin if not set
	if cfg.BotOwnerPhone == "" && len(adminPhones) > 0 {
		cfg.BotOwnerPhone = adminPhones[0]
	}

	if cfg.PairingPhone != "" {
		// Clean the pairing phone just in case
		cleanPairing := strings.ReplaceAll(cfg.PairingPhone, " ", "")
		cleanPairing = strings.ReplaceAll(cleanPairing, "+", "")
		cleanPairing = strings.ReplaceAll(cleanPairing, "-", "")

		// Ensure it's not already in the list
		exists := false
		for _, p := range adminPhones {
			if p == cleanPairing {
				exists = true
				break
			}
		}
		if !exists {
			adminPhones = append(adminPhones, cleanPairing)
		}
		cfg.PairingPhone = cleanPairing
	}

	cfg.AdminPhones = adminPhones

	if cfg.WhatsAppSessionPath == "" {
		cfg.WhatsAppSessionPath = filepath.Join(workDir, "session.db")
	}

	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	}

	if cfg.LogPath == "" {
		cfg.LogPath = filepath.Join(workDir, "logs")
	}
	return cfg, nil
}

func (cfg *Config) Validate() error {
	// SQLite is used locally, no DATABASE_URL needed
	if cfg.GeminiAPIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY is missing")
	}
	if cfg.CompanyPrefix == "" {
		cfg.CompanyPrefix = "AWB"
	}
	return nil
}

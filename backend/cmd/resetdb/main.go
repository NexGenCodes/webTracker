package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Try multiple paths so it works from any CWD
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load("../../.env") // fallback when run from cmd/resetdb/

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DIRECT_URL")
	}
	if dbURL == "" {
		log.Fatal("DATABASE_URL / DIRECT_URL is not set in backend/.env")
	}

	// Ensure search_path=public for Neon compatibility (mirrors config.Load behavior)
	if !strings.Contains(dbURL, "search_path=public") {
		if strings.Contains(dbURL, "?") {
			dbURL += "&search_path=public"
		} else {
			dbURL += "?search_path=public"
		}
	}

	// SAFETY GUARD: Never run this against production unless explicitly forced
	isProd := os.Getenv("APP_ENV") == "production" || strings.Contains(dbURL, "aws") || strings.Contains(dbURL, "rds")
	forceWipe := os.Getenv("FORCE_PROD_WIPE") == "I_KNOW_WHAT_I_AM_DOING"

	if isProd && !forceWipe {
		log.Fatal("CRITICAL ERROR: Attempted to run resetdb in a production environment. If you TRULY want to wipe all production data, set the environment variable FORCE_PROD_WIPE=I_KNOW_WHAT_I_AM_DOING")
	} else if isProd && forceWipe {
		fmt.Println("⚠️  WARNING: Production wipe override detected. Dropping all production data in 5 seconds...")
		fmt.Println("Press Ctrl+C NOW to abort.")
		
		time.Sleep(1 * time.Second)
		fmt.Println("5...")
		time.Sleep(1 * time.Second)
		fmt.Println("4...")
		time.Sleep(1 * time.Second)
		fmt.Println("3...")
		time.Sleep(1 * time.Second)
		fmt.Println("2...")
		time.Sleep(1 * time.Second)
		fmt.Println("1...")
		fmt.Println("Wiping database...")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer pool.Close()

	// Drop tables safely (in proper order to respect constraints if any)
	drops := []string{
		"DROP TABLE IF EXISTS payments CASCADE;",
		"DROP TABLE IF EXISTS telemetry CASCADE;",
		"DROP TABLE IF EXISTS Shipment CASCADE;",
		"DROP TABLE IF EXISTS GroupAuthority CASCADE;",
		"DROP TABLE IF EXISTS UserPreference CASCADE;",
		"DROP TABLE IF EXISTS SystemConfig CASCADE;",
		"DROP TABLE IF EXISTS companies CASCADE;",
		"DROP TABLE IF EXISTS schema_migrations CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_device CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_sessions CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_identity CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_identities CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_contacts CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_groups CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_prekeys CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_sender_keys CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_versions CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_version CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_identity_keys CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_pre_keys CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_app_state_sync_keys CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_app_state_version CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_app_state_mutation_macs CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_chat_settings CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_message_secrets CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_privacy_tokens CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_lid_map CASCADE;",
		"DROP TABLE IF EXISTS whatsmeow_event_buffer CASCADE;",
		"DROP TABLE IF EXISTS sessions CASCADE;",
	}
	for _, q := range drops {
		_, err = pool.Exec(context.Background(), q)
		if err != nil {
			log.Printf("Warning dropping table: %v", err)
		}
	}

	// Read schema - try multiple paths
	schemaPath := "sql/schema.sql"
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		schema, err = os.ReadFile("../../sql/schema.sql")
		if err != nil {
			log.Fatalf("Failed to read schema.sql: %v", err)
		}
	}

	_, err = pool.Exec(context.Background(), string(schema))
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	fmt.Println("Database reset successfully with pristine SQLC Schema.")

	// Auto-run generate to ensure Go code is in sync
	fmt.Println("Auto-running code generation...")
	root := findRoot()
	if root != "" {
		cmd := exec.Command("go", "run", "cmd/generate/main.go")
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("Warning: Auto-generate failed: %v", err)
		}
	} else {
		log.Println("Warning: Could not find backend root for auto-generation")
	}
}

func findRoot() string {
	// Try CWD first
	if _, err := os.Stat("sqlc.yaml"); err == nil {
		abs, _ := filepath.Abs(".")
		return abs
	}
	// Try parent dirs
	dir, _ := os.Getwd()
	for dir != "" && dir != filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "sqlc.yaml")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

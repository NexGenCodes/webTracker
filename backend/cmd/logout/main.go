package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Try multiple paths so it works from any CWD
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DIRECT_URL")
	}
	if dbURL == "" {
		log.Fatal("DATABASE_URL / DIRECT_URL is not set")
	}

	// Ensure search_path=public for Neon compatibility (mirrors config.Load behavior)
	if !strings.Contains(dbURL, "search_path=public") {
		if strings.Contains(dbURL, "?") {
			dbURL += "&search_path=public"
		} else {
			dbURL += "?search_path=public"
		}
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer pool.Close()

	// Clear WhatsApp sessions to force a new pairing code
	drops := []string{
		"DELETE FROM whatsmeow_device;",
		"DELETE FROM whatsmeow_sessions;",
	}
	for _, q := range drops {
		_, err = pool.Exec(context.Background(), q)
		if err != nil {
			log.Printf("Warning clearing session table: %v", err)
		}
	}

	fmt.Println("Logout complete. Restart the backend to generate a new pairing code.")
}

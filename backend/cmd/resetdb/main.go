package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer pool.Close()

	// Drop tables safely
	drops := []string{
		"DROP TABLE IF EXISTS Shipment CASCADE;",
		"DROP TABLE IF EXISTS GroupAuthority CASCADE;",
		"DROP TABLE IF EXISTS UserPreference CASCADE;",
		"DROP TABLE IF EXISTS SystemConfig CASCADE;",
		"DROP TABLE IF EXISTS country_timezones CASCADE;", // from old schema
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
}

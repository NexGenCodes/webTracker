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
	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DIRECT_URL")
	}
	if dbURL == "" {
		log.Fatal("DATABASE_URL / DIRECT_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer pool.Close()

	// Read schema
	schemaPath := "sql/schema.sql"
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		schema, err = os.ReadFile("../../sql/schema.sql")
		if err != nil {
			log.Fatalf("Failed to read schema.sql: %v", err)
		}
	}

	// Apply schema (CREATE IF NOT EXISTS — safe for existing tables)
	_, err = pool.Exec(context.Background(), string(schema))
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration applied successfully (schema.sql).")
}

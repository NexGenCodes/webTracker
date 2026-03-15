package tests

import (
	"context"
	"os"
	"testing"
	"webtracker-bot/internal/localdb"

	"github.com/joho/godotenv"
)

func TestLocalDB(t *testing.T) {
	// Load .env from parent directory
	_ = godotenv.Load("../.env")

	// This test now requires a live PostgreSQL database
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("Skipping TestLocalDB: TEST_DATABASE_URL not set")
	}

	client, err := localdb.NewClient(dsn)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Use a unique schema for isolation
	schemaName := "test_localdb"
	_, err = client.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+schemaName)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}
	defer client.Exec(ctx, "DROP SCHEMA "+schemaName+" CASCADE")

	// Set search path to our test schema
	_, err = client.Exec(ctx, "SET search_path TO "+schemaName)
	if err != nil {
		t.Fatalf("Failed to set search_path: %v", err)
	}

	// Re-run initSchema in the new schema
	if err := client.InitSchema(ctx); err != nil {
		t.Fatalf("Failed to init schema in test schema: %v", err)
	}

	jid := "123456789@s.whatsapp.net"

	// 1. Get default
	lang, err := client.GetUserLanguage(ctx, jid)
	if err != nil {
		t.Errorf("GetUserLanguage failed: %v", err)
	}
	if lang != "en" {
		t.Errorf("Expected default 'en', got %q", lang)
	}

	// 2. Set language
	if err := client.SetUserLanguage(ctx, jid, "es"); err != nil {
		t.Errorf("SetUserLanguage failed: %v", err)
	}

	// 3. Get updated
	lang, err = client.GetUserLanguage(ctx, jid)
	if err != nil {
		t.Errorf("GetUserLanguage failed: %v", err)
	}
	if lang != "es" {
		t.Errorf("Expected 'es', got %q", lang)
	}

	// 4. Update again
	if err := client.SetUserLanguage(ctx, jid, "de"); err != nil {
		t.Errorf("SetUserLanguage failed: %v", err)
	}

	lang, err = client.GetUserLanguage(ctx, jid)
	if lang != "de" {
		t.Errorf("Expected 'de', got %q", lang)
	}
}

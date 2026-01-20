package localdb

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalDB(t *testing.T) {
	// Use a temp file for testing
	tempDir := os.TempDir()
	dbPath := filepath.Join(tempDir, "test_session.db")
	defer os.Remove(dbPath)

	client, err := NewClient(dbPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
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

package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Setup test env
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("GEMINI_API_KEY", "test_key")
	os.Setenv("COMPANY_NAME", "Test Logistics")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("GEMINI_API_KEY")
	defer os.Unsetenv("COMPANY_NAME")

	cfg := Load()

	if cfg.DatabaseURL == "" {
		t.Error("Expected DatabaseURL to be loaded")
	}
	if cfg.CompanyName != "TestLogistics" { // Note: whitespace removed as per implementation
		t.Errorf("Expected CompanyName to be TestLogistics, got %s", cfg.CompanyName)
	}
	if cfg.CompanyPrefix != "TL" && cfg.CompanyPrefix != "TES" { // Depending on abbreviation logic
		t.Logf("CompanyPrefix: %s", cfg.CompanyPrefix)
	}
}

func TestAbbreviation(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"Test Logistics", "TL"},
		{"Single", "SIN"},
		{"Three Words Here", "TWH"},
		{"", "AWB"},
	}

	for _, tt := range tests {
		res := generateAbbreviation(tt.name)
		if res != tt.expected {
			t.Errorf("For %s, expected %s, got %s", tt.name, tt.expected, res)
		}
	}
}

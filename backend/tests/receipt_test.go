package tests

import (
	"os"
	"testing"
	"time"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/utils"
)

func TestRenderReceipt(t *testing.T) {
	// Initialize fonts (required for rendering)
	if err := utils.InitReceiptRenderer(); err != nil {
		t.Fatalf("Failed to init renderer: %v", err)
	}

	// Use real config
	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	ship := shipment.Shipment{
		TrackingID:     "AWB-TEST-12345",
		SenderName:     "Sender Name",
		RecipientName:  "Receiver Name",
		RecipientPhone: "+1234567890",
		Destination:    "123 Test St, Test City, Test Country",
		RecipientEmail: "test@example.com",
		Origin:         "Country B",
		CreatedAt:      time.Now(),
	}

	imgBytes, err := utils.RenderReceipt(ship, cfg.CompanyName, i18n.EN)
	if err != nil {
		t.Fatalf("RenderReceipt failed: %v", err)
	}

	if len(imgBytes) == 0 {
		t.Fatal("RenderReceipt returned empty bytes")
	}

	// Save to file for manual inspection
	err = os.WriteFile("receipt_test_output.jpg", imgBytes, 0644)
	if err != nil {
		t.Errorf("Failed to write output file: %v", err)
	}
}

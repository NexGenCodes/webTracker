package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/utils"
)

func TestRenderReceipts(t *testing.T) {
	// 1. Setup mock shipment
	now := time.Now().UTC()
	departure := now.Add(24 * time.Hour)
	arrival := now.Add(72 * time.Hour)

	mockShipment := shipment.Shipment{
		TrackingID:           "TEST-ABC-123",
		Status:               shipment.StatusPending,
		SenderName:           "John Doe",
		SenderPhone:          "+1 555-0199",
		Origin:               "London, UK",
		RecipientName:        "Jane Smith",
		RecipientPhone:       "+1 555-0188",
		RecipientEmail:       "jane.smith@example.com",
		RecipientAddress:     "123 Maple Avenue,\nEvergreen Park,\nChicago, IL 60642",
		Destination:          "Chicago, USA",
		CargoType:            "DIPLOMATIC BOX",
		Weight:               12.50,
		ScheduledTransitTime: &departure,
		ExpectedDeliveryTime: &arrival,
		SenderTimezone:       "Europe/London",
		RecipientTimezone:    "America/Chicago",
	}

	// 2. Initialize renderer (with fonts path)
	err := utils.InitReceiptRenderer(false) // Sets useOptimized = false initially
	if err != nil {
		t.Fatalf("Failed to initialize receipt renderer: %v", err)
	}

	// Create test_output directory if not exists
	testDir := "test_output"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// 3. Test Legacy Render
	t.Run("Legacy Render", func(t *testing.T) {
		bytes, err := utils.RenderReceiptLegacy(mockShipment, "Global Express", i18n.EN)
		if err != nil {
			t.Fatalf("Legacy render failed: %v", err)
		}
		if len(bytes) == 0 {
			t.Fatal("Legacy render returned 0 bytes")
		}

		path := filepath.Join(testDir, "legacy_receipt.jpg")
		if err := os.WriteFile(path, bytes, 0644); err != nil {
			t.Fatalf("Failed to write legacy file: %v", err)
		}
		t.Logf("Legacy receipt saved to %s", path)
	})

	// 4. Test Optimized Render
	t.Run("Optimized Render", func(t *testing.T) {
		// Initialize for optimized
		err := utils.InitReceiptRenderer(true)
		if err != nil {
			t.Fatalf("Failed to re-init for optimized: %v", err)
		}

		bytes, err := utils.RenderReceiptOptimized(mockShipment, "Global Express", i18n.EN)
		if err != nil {
			t.Fatalf("Optimized render failed: %v", err)
		}
		if len(bytes) == 0 {
			t.Fatal("Optimized render returned 0 bytes")
		}

		path := filepath.Join(testDir, "optimized_receipt.jpg")
		if err := os.WriteFile(path, bytes, 0644); err != nil {
			t.Fatalf("Failed to write optimized file: %v", err)
		}
		t.Logf("Optimized receipt saved to %s", path)
	})
}

package utils

import (
	"os"
	"testing"
	"time"
	"webtracker-bot/internal/models"
)

func TestRenderReceipt(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	// Initialize fonts
	if err := InitReceiptRenderer(); err != nil {
		t.Fatalf("Failed to init renderer: %v", err)
	}

	shipment := models.Shipment{
		TrackingNumber:  "AWB-TEST-12345",
		SenderName:      "odo CHINAGOROM",
		ReceiverName:    "EZE EMMANUEL CHINAGOROM",
		ReceiverPhone:   "+1234567890",
		ReceiverAddress: "123 Test St, Test City, Test Country",
		ReceiverCountry: "colombia",
		ReceiverEmail:   "okechukwuzealtherealdeal@gmail.com",
		SenderCountry:   "afganistan",
		CreatedAt:       time.Now(),
	}

	imgBytes, err := RenderReceipt(shipment)
	if err != nil {
		t.Fatalf("RenderReceipt failed: %v", err)
	}

	if len(imgBytes) == 0 {
		t.Fatal("RenderReceipt returned empty bytes")
	}

	// Save to file for manual inspection
	err = os.WriteFile("receipt_test_output.jpg", imgBytes, 0644)
	if err != nil {
		t.Fatalf("Failed to write output file: %v", err)
	}

	t.Logf("Successfully generated receipt_test_output.jpg (%d bytes)", len(imgBytes))

}

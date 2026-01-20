package tests

import (
	"strings"
	"testing"
	"time"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/utils"
)

func TestGenerateWaybill(t *testing.T) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	shipment := models.Shipment{
		TrackingNumber:  "AWB-TEST-123",
		Status:          "PENDING",
		SenderName:      "Global Sender Org",
		ReceiverName:    "Local Receiver Person",
		ReceiverPhone:   "0800-WAYBILL-00",
		ReceiverAddress: "456 Enterprise Way, District 9",
		ReceiverCountry: "Nigeria",
		CreatedAt:       time.Now(),
	}

	got := utils.GenerateWaybill(shipment, cfg.CompanyName)

	// Check for key elements and formatting
	checks := []string{
		"AWB-TEST-123",
		"Global Sender Org",
		"Local Receiver Person",
		"456 Enterprise Way",
		strings.ToUpper(cfg.CompanyName),
		"━━", // Check for block borders
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("Waybill missing required element: %s", check)
		}
	}
}

func TestGenerateWaybillEmptyFields(t *testing.T) {
	cfg, _ := config.LoadFromEnv()
	companyName := "EMPTY CORP"
	if cfg != nil {
		companyName = cfg.CompanyName
	}

	shipment := models.Shipment{
		TrackingNumber: "EMPTY-TEST",
	}

	got := utils.GenerateWaybill(shipment, companyName)
	if !strings.Contains(got, "EMPTY-TEST") {
		t.Error("Waybill should at least contain the tracking number even if other fields are empty")
	}
}

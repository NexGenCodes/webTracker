package tests

import (
	"strings"
	"testing"
	"time"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/utils"
)

func TestGenerateWaybill(t *testing.T) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	ship := shipment.Shipment{
		TrackingID:     "AWB-TEST-123",
		Status:         "PENDING",
		SenderName:     "Global Sender Org",
		RecipientName:  "Local Receiver Person",
		RecipientPhone: "0800-WAYBILL-00",
		Destination:    "456 Enterprise Way, District 9",
		Origin:         "Nigeria",
		CreatedAt:      time.Now(),
	}

	got := utils.GenerateWaybill(ship, cfg.CompanyName)

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

	ship := shipment.Shipment{
		TrackingID: "EMPTY-TEST",
	}

	got := utils.GenerateWaybill(ship, companyName)
	if !strings.Contains(got, "EMPTY-TEST") {
		t.Error("Waybill should at least contain the tracking number even if other fields are empty")
	}
}

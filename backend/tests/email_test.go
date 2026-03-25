package tests

import (
	"testing"
	"time"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/shipment"
)

func TestSendDeliveryEmail(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "", // Empty SMTPHost should cause the function to return early without error
		SMTPUsername: "",
		CompanyName:  "Test Express",
	}

	expectedDelivery := time.Now().Add(24 * time.Hour)
	s := &shipment.Shipment{
		TrackingID:           "TEST-12345",
		RecipientName:        "John Doe",
		RecipientEmail:       "john@example.com",
		RecipientAddress:     "123 Main St",
		Destination:          "Lagos, Nigeria",
		ExpectedDeliveryTime: &expectedDelivery,
	}

	// This should not panic or error if SMTP is not configured
	notif.SendDeliveryEmail(cfg, s)
}

package tests

import (
	"testing"
	"time"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/shipment"
)

func TestSendShipmentEmail(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "", // Empty SMTPHost should cause the function to return early without error
		SMTPUsername: "",
		CompanyName:  "Test Express",
	}

	s := &shipment.Shipment{
		TrackingID:           "TEST-12345",
		RecipientName:        "John Doe",
		RecipientEmail:       "john@example.com",
		RecipientAddress:     "123 Main St",
		Destination:          "Lagos, Nigeria",
		ExpectedDeliveryTime: time.Now().Add(24 * time.Hour),
	}

	// This should not panic or error if SMTP is not configured
	notif.SendShipmentEmail(cfg, s, "http://localhost:3000/track/TEST-12345")
}

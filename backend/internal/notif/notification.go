package notif

import (
	"strings"
	"time"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/shipment"
)

// ---------------------------------------------------------------------------
// Package-level helpers for callers that don't yet inject a Mailer.
// Prefer injecting *Mailer via DI for new code.
// ---------------------------------------------------------------------------



// SendSetupLinkEmail sends a magic setup link to a company admin
func SendSetupLinkEmail(cfg *config.Config, adminEmail, companyName, setupToken string) {
	m := NewMailer(cfg)
	companyName = strings.ToUpper(companyName)
	if companyName == "" {
		companyName = "AIRWAYBILL"
	}
	e := SetupLinkEmail(adminEmail, companyName, cfg.FrontendURL, setupToken)
	m.Send(e)
}

// SendDeliveryEmail sends a professional email when a shipment is delivered
func SendDeliveryEmail(cfg *config.Config, s *shipment.Shipment, companyName string) {
	m := NewMailer(cfg)
	companyName = strings.ToUpper(companyName)
	if companyName == "" {
		companyName = "AIRWAYBILL"
	}
	e := DeliveryEmail(
		s.RecipientEmail,
		s.RecipientName,
		s.TrackingID,
		companyName,
		time.Now().Format("January 02, 2006"),
	)
	m.Send(e)
}

// SendOTPEmail sends a 6-digit verification code email
func SendOTPEmail(cfg *config.Config, email, otp string) {
	m := NewMailer(cfg)
	e := OTPEmail(email, otp)
	m.Send(e)
}

package notif

import (
	"strings"
	"sync"
	"time"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/i18n"
)

// ---------------------------------------------------------------------------
// Package-level singleton mailer.
// Call InitMailer(cfg) once at startup and ShutdownMailer() at teardown.
// ---------------------------------------------------------------------------

var (
	globalMailer *Mailer
	mailerOnce   sync.Once
)

// InitMailer initializes the package-level mailer singleton. Safe to call multiple times.
func InitMailer(cfg *config.Config) *Mailer {
	mailerOnce.Do(func() {
		globalMailer = NewMailer(cfg)
	})
	return globalMailer
}

// ShutdownMailer gracefully stops the singleton mailer's worker pool.
func ShutdownMailer() {
	if globalMailer != nil {
		globalMailer.Shutdown()
	}
}

// getMailer returns the singleton or creates an ephemeral one as a fallback.
func getMailer(cfg *config.Config) *Mailer {
	if globalMailer != nil {
		return globalMailer
	}
	// Fallback: direct synchronous send (no worker pool leak)
	return &Mailer{cfg: cfg}
}

// SendSetupLinkEmail sends a magic setup link to a company admin
func SendSetupLinkEmail(cfg *config.Config, adminEmail, companyName, setupToken string) {
	m := getMailer(cfg)
	companyName = strings.ToUpper(companyName)
	if companyName == "" {
		companyName = "AIRWAYBILL"
	}
	e := SetupLinkEmail(adminEmail, companyName, cfg.FrontendURL, setupToken)
	m.SendAsync(e)
}

// SendDeliveryEmail sends a professional email when a shipment is delivered
func SendDeliveryEmail(cfg *config.Config, email, trackingID, recipientName, companyName, destCountry string) {
	m := getMailer(cfg)
	companyName = strings.ToUpper(companyName)
	if companyName == "" {
		companyName = "AIRWAYBILL"
	}
	lang := i18n.GetLanguageForCountry(destCountry)
	e := DeliveryEmail(
		email,
		recipientName,
		trackingID,
		companyName,
		time.Now().Format(i18n.GetDateFormat(lang)),
		lang,
	)
	m.SendAsync(e)
}

// SendOTPEmail sends a 6-digit verification code email
func SendOTPEmail(cfg *config.Config, email, otp string) {
	m := getMailer(cfg)
	e := OTPEmail(email, otp)
	m.SendAsync(e)
}

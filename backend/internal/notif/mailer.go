package notif

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
)

// Mailer is the reusable, app-wide email sender.
// Initialize once with NewMailer(cfg) and inject where needed.
type Mailer struct {
	cfg *config.Config
}

// NewMailer creates a mailer from the app config.
func NewMailer(cfg *config.Config) *Mailer {
	return &Mailer{cfg: cfg}
}

// Email represents a single outbound email.
type Email struct {
	To       string
	Subject  string
	HTMLBody string
	FromName string // optional override; defaults to "WebTracker"
}

// Send dispatches an email. Safe to call from a goroutine.
func (m *Mailer) Send(e Email) {
	cfg := m.cfg
	if cfg.SMTPHost == "" || cfg.SMTPUsername == "" || e.To == "" {
		logger.Warn().Str("to", e.To).Msg("mailer: skipping send — SMTP not configured or missing recipient")
		return
	}

	fromName := e.FromName
	if fromName == "" {
		fromName = "WebTracker"
	}
	fromAddr := cfg.NotifyEmail
	if fromAddr == "" {
		fromAddr = cfg.SMTPUsername
	}

	raw := "From: " + fromName + " <" + fromAddr + ">\r\n" +
		"To: " + e.To + "\r\n" +
		"Subject: " + e.Subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		e.HTMLBody

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)

	var err error
	if cfg.SMTPPort == 465 {
		err = smtps(addr, cfg.SMTPHost, cfg.SMTPUsername, cfg.SMTPPassword, fromAddr, e.To, []byte(raw))
	} else {
		auth := smtp.PlainAuth("",
			strings.TrimSpace(cfg.SMTPUsername),
			strings.TrimSpace(cfg.SMTPPassword),
			strings.TrimSpace(cfg.SMTPHost),
		)
		err = smtp.SendMail(addr, auth, fromAddr, []string{e.To}, []byte(raw))
	}

	if err != nil {
		logger.Error().Err(err).Str("to", e.To).Str("subject", e.Subject).Msg("mailer: send failed")
	} else {
		logger.Info().Str("to", e.To).Str("subject", e.Subject).Msg("mailer: delivered")
	}
}

// SendAsync dispatches the email on a new goroutine.
func (m *Mailer) SendAsync(e Email) {
	go m.Send(e)
}

// ---------------------------------------------------------------------------
// SMTP/S transport (shared, private)
// ---------------------------------------------------------------------------

func smtps(addr, host, user, pass, from, to string, msg []byte) error {
	user = strings.TrimSpace(user)
	pass = strings.TrimSpace(pass)
	host = strings.TrimSpace(host)

	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Quit()

	if err = c.Auth(smtp.PlainAuth("", user, pass, host)); err != nil {
		return err
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		return err
	}
	return w.Close()
}

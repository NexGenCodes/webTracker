package notif

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
)

// SendPairingCodeEmail sends the pairing code via a professional HTML email
func SendPairingCodeEmail(cfg *config.Config, code string) {
	if cfg.SMTPHost == "" || cfg.SMTPUsername == "" || cfg.NotifyEmail == "" {
		return
	}

	companyName := strings.ToUpper(cfg.CompanyName)
	if companyName == "" {
		companyName = "AirWayBill"
	}

	subject := fmt.Sprintf("[%s] WhatsApp Pairing Code", companyName)
	recipient := cfg.NotifyEmail
	pairingPhone := cfg.PairingPhone

	// HTML Template
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
        <h1 style="color: #007bff; font-size: 24px; margin-bottom: 10px;">%s</h1>
        <h2 style="color: #333; font-size: 18px; margin-bottom: 25px;">WhatsApp Integration Setup</h2>
        
        <p style="color: #555; line-height: 1.6;">You have requested a pairing code to link your WhatsApp account to the tracking bot.</p>
        
        <div style="background-color: #f8f9fa; border-left: 4px solid #007bff; padding: 15px; margin: 25px 0;">
            <p style="margin: 0; color: #333;"><strong>Target Phone:</strong> %s</p>
        </div>

        <div style="text-align: center; margin: 35px 0; padding: 25px; background: #e9ecef; border-radius: 8px;">
            <p style="margin: 0 0 10px 0; font-size: 14px; text-transform: uppercase; letter-spacing: 1px; color: #666;">Your Pairing Code</p>
            <div style="font-size: 42px; font-weight: bold; letter-spacing: 8px; color: #000; font-family: 'Courier New', Courier, monospace;">%s</div>
        </div>

        <p style="color: #dc3545; font-size: 14px; font-weight: bold; text-align: center;">⏱️ THIS CODE EXPIRES IN 2 MINUTES</p>

        <p style="margin-top: 30px; border-top: 1px solid #eee; padding-top: 20px; color: #888; font-size: 12px;">
            If you did not request this code, please ignore this email or contact security.
            <br>&copy; 2026 %s. All rights reserved.
        </p>
    </div>
</body>
</html>`, companyName, pairingPhone, code, companyName)

	msg := "From: " + cfg.SMTPUsername + "\r\n" +
		"To: " + recipient + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		body

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	logger.Info().Str("to", recipient).Int("port", cfg.SMTPPort).Msg("Attempting to send pairing email")

	var err error
	if cfg.SMTPPort == 465 {
		// SMTPS (Direct TLS)
		err = sendSMTPS(addr, cfg.SMTPHost, cfg.SMTPUsername, cfg.SMTPPassword, recipient, []byte(msg))
	} else {
		// Standard SMTP (STARTTLS)
		user := strings.TrimSpace(cfg.SMTPUsername)
		pass := strings.TrimSpace(cfg.SMTPPassword)
		host := strings.TrimSpace(cfg.SMTPHost)
		auth := smtp.PlainAuth("", user, pass, host)
		err = smtp.SendMail(addr, auth, user, []string{recipient}, []byte(msg))
	}

	if err != nil {
		logger.Error().Err(err).Msg("Failed to send pairing code email")
	} else {
		logger.Info().Str("email", recipient).Msg("Professional pairing email delivered")
	}
}

func sendSMTPS(addr, host, user, pass, to string, msg []byte) error {
	user = strings.TrimSpace(user)
	pass = strings.TrimSpace(pass)
	host = strings.TrimSpace(host)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Quit()

	auth := smtp.PlainAuth("", user, pass, host)
	if err = c.Auth(auth); err != nil {
		return err
	}

	if err = c.Mail(user); err != nil {
		return err
	}

	if err = c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	return w.Close()
}

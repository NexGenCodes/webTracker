package notif

import (
	"context"
	"fmt"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/shipment"

	"go.mau.fi/whatsmeow/types"
)

// SendStatusAlert sends a WhatsApp and/or email alert when a shipment transitions.
// It uses the WhatsAppSender interface to ensure messages are enqueued with jitter and backpressure.
func SendStatusAlert(ctx context.Context, sender models.WhatsAppSender, cfg *config.Config, companyName, jidStr, tracking, status, email, destCountry string) {
	if jidStr == "" || sender == nil {
		return
	}
	jid, err := types.ParseJID(jidStr)
	if err != nil {
		logger.Warn().Str("jid", jidStr).Msg("Failed to parse JID for status alert")
		return
	}

	var msg string
	link := ""
	if cfg != nil && cfg.FrontendURL != "" {
		link = fmt.Sprintf("\n🌐 *Track Here:* %s/track/%s", cfg.FrontendURL, tracking)
	}

	switch status {
	case shipment.StatusIntransit:
		msg = fmt.Sprintf("✈️ *SHIPMENT UPDATE*\n\nTracking ID: *%s*\nStatus: *IN TRANSIT*\n\nYour shipment has securely departed the origin facility and is now en route to the destination country.%s", tracking, link)
	case "outfordelivery":
		msg = fmt.Sprintf("🚚 *OUT FOR DELIVERY*\n\nTracking ID: *%s*\nStatus: *OUT FOR DELIVERY*\n\nYour shipment is with our local courier and will be delivered to the recipient address today. Please ensure someone is available to receive it.%s", tracking, link)
	case shipment.StatusDelivered:
		msg = fmt.Sprintf("🛬 *NOTICE OF ARRIVAL*\n\nTracking ID: *%s*\nStatus: *ARRIVED AT DESTINATION*\n\nYour shipment has successfully arrived in the destination country and is securely held at our facility. A regional agent will contact the recipient shortly.%s", tracking, link)

		if email != "" && cfg != nil {
			SendDeliveryEmail(cfg, email, tracking, "", companyName, destCountry)
		}
	default:
		return
	}

	// Ensure we send to the bare JID (all devices)
	bareJid := types.JID{User: jid.User, Server: jid.Server}

	// sender.Send automatically adds the bot footer and enqueues to the jitter queue
	sender.Send(bareJid, msg)
	logger.Info().Str("chat", jidStr).Str("status", status).Msg("Status alert enqueued")
}

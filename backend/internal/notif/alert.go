package notif

import (
	"context"
	"fmt"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/shipment"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

// SendStatusAlert sends a WhatsApp and/or email alert when a shipment transitions.
func SendStatusAlert(ctx context.Context, wa *whatsmeow.Client, cfg *config.Config, jidStr, tracking, status, email string) {
	if jidStr == "" {
		return
	}
	jid, err := types.ParseJID(jidStr)
	if err != nil {
		logger.Warn().Str("jid", jidStr).Msg("Failed to parse JID for status alert")
		return
	}

	var msg string
	switch status {
	case shipment.StatusIntransit:
		msg = fmt.Sprintf("✈️ *SHIPMENT UPDATE*\n\nTracking ID: *%s*\nStatus: *IN TRANSIT*\n\nYour shipment has securely departed the origin facility and is now en route to the destination country.", tracking)
	case shipment.StatusDelivered:
		msg = fmt.Sprintf("🛬 *NOTICE OF ARRIVAL*\n\nTracking ID: *%s*\nStatus: *ARRIVED AT DESTINATION*\n\nYour shipment has successfully arrived in the destination country and is securely held at our facility. A regional agent will contact the recipient shortly.", tracking)
		
		if email != "" && cfg != nil {
			SendDeliveryEmail(cfg, &shipment.Shipment{
				TrackingID:     tracking,
				RecipientEmail: email,
			})
		}
	default:
		return
	}

	// Add Bot Footer
	msg += "\n\n_🤖Bot_"

	content := &waProto.Message{
		Conversation: models.StrPtr(msg),
	}

	_, err = wa.SendMessage(ctx, jid, content)
	if err != nil {
		logger.Error().Err(err).Str("chat", jidStr).Msg("Failed to send status alert")
	} else {
		logger.Info().Str("chat", jidStr).Str("status", status).Msg("Status alert sent")
	}
}

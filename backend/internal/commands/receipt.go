package commands

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/receipt"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/utils"
)

// ReceiptHandler handles !receipt [ID]
type ReceiptHandler struct {
	Sender models.WhatsAppSender
}

func (h *ReceiptHandler) Execute(ctx context.Context, shipUC models.ShipmentUsecase, configUC models.ConfigUsecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	if !isAdmin {
		return Result{Message: "🔒 *ACCESS DENIED*\nOnly admins can regenerate receipts."}
	}

	var trackingID string
	jid := utils.GetJID(ctx)

	if len(args) < 1 {
		var err error
		trackingID, err = shipUC.GetLastForUser(ctx, companyID, jid)
		if err != nil || trackingID == "" {
			return Result{Message: "🧾 *RECEIPT REGENERATION*\n\nUsage: `!receipt [TrackingID]`"}
		}
	} else {
		trackingID = strings.ToUpper(args[0])
	}

	// Fetch shipment to get JID
	dbShip, err := shipUC.Track(ctx, companyID, trackingID)
	if err != nil || dbShip == nil {
		return Result{Message: "❌ *NOT FOUND*\nCould not find a shipment with that ID."}
	}

	// Map to Domain Model for Rendering
	s := shipment.ToDomain(*dbShip)

	// Render synchronous
	receiptImg, err := receipt.RenderReceipt(s, h.Sender.GetCompanyName(), i18n.Language(lang))
	if err != nil {
		return Result{Message: "❌ *RENDER FAILED*", Error: err}
	}

	return Result{
		Message: "", // Binary response handled by worker
		Image:   receiptImg,
	}
}

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/receipt"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/utils"
)

// InfoHandler handles !info [ID]
type InfoHandler struct {
	CompanyName   string
	CompanyPrefix string
}

func (h *InfoHandler) Execute(ctx context.Context, shipUC models.ShipmentUsecase, configUC models.ConfigUsecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	var trackingID string
	jid := utils.GetJID(ctx)

	if len(args) < 1 {
		// Attempt contextual lookup
		var err error
		trackingID, err = shipUC.GetLastForUser(ctx, companyID, jid)
		if err != nil || trackingID == "" {
			company := strings.ToUpper(h.CompanyName)
			if company == "" {
				company = "COMMAND"
			}
			msg := fmt.Sprintf("🚀 *%s COMMAND CENTER*\n\n", company) +
				"━━━━━━━━━━━━━━━━━━━━━━━\n"

			if isAdmin {
				msg += "1️⃣ `!stats` - Daily Operations Summary\n"
			}

			msg += "2️⃣ `!info [TrackingID]` - Shipment Information Tracker\n" +
				"━━━━━━━━━━━━━━━━━━━━━━━\n\n" +
				"*PRO TIP:*\n" +
				fmt.Sprintf("_Use `!info %s-123456789` for full details._", h.CompanyPrefix)
			return Result{Message: msg}
		}
	} else {
		trackingID = strings.ToUpper(args[0])
	}

	dbShip, err := shipUC.Track(ctx, companyID, trackingID)
	if err != nil {
		return Result{Message: i18n.T(i18nLang(lang), "ERR_DB_ERROR"), Error: err}
	}

	if dbShip == nil {
		return Result{Message: i18n.T(i18nLang(lang), "ERR_NOT_FOUND")}
	}

	// Map DB model to Domain model for waybill generation
	s := shipment.ToDomain(*dbShip)

	wb := receipt.GenerateWaybill(s, h.CompanyName)
	return Result{Message: "```\n" + wb + "\n```"}
}

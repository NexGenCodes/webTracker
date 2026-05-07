package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"webtracker-bot/internal/models"
	"webtracker-bot/internal/utils"
)

// DeleteHandler handles !delete [trackingID]
type DeleteHandler struct{}

func (h *DeleteHandler) Execute(ctx context.Context, shipUC models.ShipmentUsecase, configUC models.ConfigUsecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	var trackingID string
	jid := utils.GetJID(ctx)

	if len(args) < 1 {
		var err error
		trackingID, err = shipUC.GetLastForUser(ctx, companyID, jid)
		if err != nil || trackingID == "" {
			return Result{Message: "🗑️ *DELETE SHIPMENT*\n\nUsage: `!delete [TrackingID]`"}
		}
	} else {
		trackingID = strings.ToUpper(args[0])
	}

	err := shipUC.Delete(ctx, companyID, trackingID)
	if err != nil {
		return Result{Message: fmt.Sprintf("❌ *DELETE FAILED*\n_%v_", err)}
	}

	return Result{Message: fmt.Sprintf("🗑️ *SHIPMENT DELETED*\n\nThe shipment *%s* has been permanently removed.", trackingID)}
}

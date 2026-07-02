package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"webtracker-bot/internal/models"
)

// DeleteHandler handles !delete [trackingID]
type DeleteHandler struct{}

func (h *DeleteHandler) Execute(ctx context.Context, shipUC models.ShipmentUsecase, configUC models.ConfigUsecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	if len(args) < 1 {
		return Result{Message: "🗑️ *DELETE SHIPMENT*\n\nUsage: `!delete [TrackingID]`"}
	}
	trackingID := strings.ToUpper(args[0])

	err := shipUC.Delete(ctx, companyID, trackingID)
	if err != nil {
		return Result{Message: fmt.Sprintf("❌ *DELETE FAILED*\n_%v_", err)}
	}

	return Result{Message: fmt.Sprintf("🗑️ *SHIPMENT DELETED*\n\nThe shipment *%s* has been permanently removed.", trackingID)}
}

package api

import (
	"context"
	"time"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/shipment"

	"github.com/google/uuid"
)

// PlanLimits maps plan_type to the maximum shipments allowed per billing cycle.
var PlanLimits = map[string]int64{
	"starter":    50,
	"pro":        250,
	"enterprise": 1000,
	"scale":      1000,
}

// IsSuperAdmin checks if the given company ID matches the configured super admin.
func IsSuperAdmin(cfg *config.Config, companyID uuid.UUID) bool {
	if cfg.SuperAdminCompanyID == "" {
		return false
	}
	superID, err := uuid.Parse(cfg.SuperAdminCompanyID)
	if err != nil {
		return false
	}
	return companyID == superID
}

// CheckShipmentCap verifies whether the company has remaining shipments in
// the current billing cycle. Returns (remaining, error).
// A non-nil error with status 402 means the cap is reached.
func CheckShipmentCap(
	ctx context.Context,
	cfg *config.Config,
	shipmentUC *shipment.Usecase,
	companyID uuid.UUID,
	planType string,
) (remaining int64, err error) {
	// Super admin is unlimited
	if IsSuperAdmin(cfg, companyID) {
		return -1, nil // -1 signals unlimited
	}

	limit, ok := PlanLimits[planType]
	if !ok {
		// Unknown plan — default to starter cap
		limit = PlanLimits["starter"]
	}

	// Count shipments created since the start of the current calendar month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	created, err := shipmentUC.CountCreatedSince(ctx, companyID, startOfMonth)
	if err != nil {
		return 0, err
	}

	remaining = limit - created
	if remaining <= 0 {
		return 0, nil // caller should return 402
	}

	return remaining, nil
}

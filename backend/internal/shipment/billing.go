package shipment

import (
	"context"
	"database/sql"
	"time"
	"webtracker-bot/internal/config"

	"github.com/google/uuid"
)

// PlanLimits maps plan_type to the maximum shipments allowed per billing cycle.
var PlanLimits = map[string]int64{
	"trial":      50,
	"starter":    50,
	"pro":        250,
	"enterprise": 1000,
}

// IsSuperAdmin checks if the given company ID matches the configured super admin.
func IsSuperAdmin(cfg *config.Config, companyID uuid.UUID) bool {
	if cfg == nil || cfg.SuperAdminCompanyID == "" {
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
func (u *Usecase) CheckShipmentCap(
	ctx context.Context,
	cfg *config.Config,
	companyID uuid.UUID,
	planType string,
	expiry sql.NullTime,
) (remaining int64, err error) {
	// Super admin is unlimited
	if IsSuperAdmin(cfg, companyID) {
		return -1, nil // -1 signals unlimited
	}

	// Check if subscription/trial has expired
	now := time.Now()
	if expiry.Valid && expiry.Time.Before(now) {
		return 0, nil // 0 remaining implies payment required / expired
	}

	limit, ok := PlanLimits[planType]
	if !ok {
		// Unknown plan — default to starter cap
		limit = PlanLimits["starter"]
	}

	// Count shipments created since the start of the current calendar month
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	created, err := u.CountCreatedSince(ctx, companyID, startOfMonth)
	if err != nil {
		return 0, err
	}

	if created >= limit {
		return 0, nil
	}

	return limit - created, nil
}

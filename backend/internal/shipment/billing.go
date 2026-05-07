package shipment

import (
	"context"
	"database/sql"
	"time"
	"webtracker-bot/internal/billing"
	"webtracker-bot/internal/config"

	"github.com/google/uuid"
)

// CheckShipmentCap verifies whether the company has remaining shipments in
// the current billing cycle. Returns (remaining, error).
// A return value of -1 signals unlimited (super admin).
func (u *Usecase) CheckShipmentCap(
	ctx context.Context,
	cfg *config.Config,
	companyID uuid.UUID,
	planType string,
	expiry sql.NullTime,
) (remaining int64, err error) {
	// Super admin is unlimited
	if billing.IsSuperAdmin(cfg, companyID) {
		return -1, nil // -1 signals unlimited
	}

	// Check if subscription/trial has expired
	now := time.Now()
	if expiry.Valid && expiry.Time.Before(now) {
		return 0, nil // 0 remaining implies payment required / expired
	}

	// Resolve the shipment cap from the single source of truth (billing.Plan)
	plan, err := billing.GetPlanByID(planType)
	if err != nil {
		// Unknown plan — default to starter cap
		plan = billing.PlanStarter
	}
	limit := plan.MaxShipments

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

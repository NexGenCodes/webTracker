package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/payment"
	"webtracker-bot/internal/shipment"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// BillingHandler serves plan data from the backend source of truth.
type BillingHandler struct {
	cfg      *config.Config
	configUC *config.Usecase
}

func NewBillingHandler(cfg *config.Config, configUC *config.Usecase) *BillingHandler {
	return &BillingHandler{cfg: cfg, configUC: configUC}
}

func (h *BillingHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/billing")
	api.Get("/plans", h.getPlans)
}

func (h *BillingHandler) getPlans(c *fiber.Ctx) error {
	ctx := c.Context()
	dbPlans, err := h.configUC.GetActivePlans(ctx)
	if err != nil || len(dbPlans) == 0 {
		// Fallback to static plans if DB fails or is empty
		return c.JSON(payment.GetPlans())
	}

	var plans []payment.Plan
	for _, p := range dbPlans {
		var features []string
		_ = json.Unmarshal(p.Features, &features)

		plans = append(plans, payment.Plan{
			ID:          p.ID,
			Name:        p.Name,
			NameKey:     p.NameKey,
			DescKey:     p.DescKey,
			Price:       int(p.BasePrice),
			Currency:    p.Currency,
			IntervalKey: p.IntervalKey,
			Popular:     p.Popular.Bool,
			TrialKey:    p.TrialKey.String,
			BtnKey:      p.BtnKey,
			Features:    features,
		})
	}
	return c.JSON(plans)
}

// PlanLimits maps plan_type to the maximum shipments allowed per billing cycle.
var PlanLimits = map[string]int64{
	"trial":      50,
	"starter":    50,
	"pro":        250,
	"enterprise": 1000,
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

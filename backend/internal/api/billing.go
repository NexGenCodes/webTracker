package api

import (
	"encoding/json"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/payment"

	"github.com/gofiber/fiber/v2"
)

// BillingHandler serves plan data from the backend source of truth.
type BillingHandler struct {
	cfg      *config.Config
	configUC models.ConfigUsecase
}

func NewBillingHandler(cfg *config.Config, configUC models.ConfigUsecase) *BillingHandler {
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



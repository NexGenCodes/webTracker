package api

import (
	"encoding/json"

	"webtracker-bot/internal/billing"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// BillingHandler serves plan data, handles subscriptions, and processes Paystack webhooks.
type BillingHandler struct {
	cfg      *config.Config
	configUC models.ConfigUsecase
	paystack *billing.PaystackService
}

func NewBillingHandler(cfg *config.Config, configUC models.ConfigUsecase) *BillingHandler {
	return &BillingHandler{
		cfg:      cfg,
		configUC: configUC,
		paystack: billing.NewPaystackService(cfg.PaystackSecretKey),
	}
}

func (h *BillingHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/billing")
	api.Get("/plans", h.getPlans)
	api.Post("/subscribe", h.subscribe)
	api.Get("/subscription-status", h.getSubscriptionStatus)

	// Webhooks — outside the /api/billing group for Paystack compatibility
	app.Post("/api/webhooks/paystack", h.paystackWebhook)
}

func (h *BillingHandler) getPlans(c *fiber.Ctx) error {
	ctx := c.Context()
	dbPlans, err := h.configUC.GetActivePlans(ctx)
	if err != nil || len(dbPlans) == 0 {
		// Fallback to static plans if DB fails or is empty
		return c.JSON(billing.GetPlans())
	}

	var plans []billing.Plan
	for _, p := range dbPlans {
		var features []string
		_ = json.Unmarshal(p.Features, &features)

		plans = append(plans, billing.Plan{
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

// subscribe initializes a Paystack transaction for the company.
func (h *BillingHandler) subscribe(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	// Super admin bypasses billing
	if billing.IsSuperAdmin(h.cfg, companyID) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Super admin accounts do not require subscriptions"})
	}

	company, err := h.configUC.GetCompanyByID(c.Context(), companyID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Company not found"})
	}

	var req struct {
		CallbackURL string `json:"callback_url"`
		Plan        string `json:"plan"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Resolve the plan price — DB first, static fallback
	var planPriceKobo int
	var planID string

	dbPlan, dbErr := h.configUC.GetPlanByID(c.Context(), req.Plan)
	if dbErr == nil {
		planPriceKobo = int(dbPlan.BasePrice)
		planID = dbPlan.ID
	} else {
		plan, err := billing.GetPlanByID(req.Plan)
		if err != nil {
			plan = billing.PlanPro // Default
		}
		planPriceKobo = plan.Price
		planID = plan.ID
	}

	metadata := map[string]interface{}{
		"company_id": companyID.String(),
		"plan":       planID,
	}

	authURL, reference, err := h.paystack.InitializeTransaction(company.AdminEmail, planPriceKobo, req.CallbackURL, metadata)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize Paystack transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start payment"})
	}

	logger.Info().
		Str("company_id", companyID.String()).
		Str("reference", reference).
		Str("plan", planID).
		Msg("Payment transaction initialized")

	return c.JSON(fiber.Map{
		"success":           true,
		"authorization_url": authURL,
		"reference":         reference,
	})
}

func (h *BillingHandler) getSubscriptionStatus(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	company, err := h.configUC.GetCompanyByID(c.Context(), companyID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Company not found"})
	}

	return c.JSON(fiber.Map{
		"status": company.SubscriptionStatus.String,
		"expiry": company.SubscriptionExpiry.Time,
		"plan":   company.PlanType.String,
	})
}

// paystackWebhook handles inbound webhook events from Paystack.
// All monetary comparisons are done in kobo (subunits) to avoid floating-point bugs.
func (h *BillingHandler) paystackWebhook(c *fiber.Ctx) error {
	signature := c.Get("x-paystack-signature")
	if signature == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	payload := c.Body()
	if !h.paystack.VerifySignature(payload, signature) {
		logger.Error().Msg("Invalid Paystack webhook signature")
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	var event struct {
		Event string `json:"event"`
		Data  struct {
			Status    string `json:"status"`
			Reference string `json:"reference"`
			Amount    int    `json:"amount"` // Paystack sends amount in kobo
			Customer  struct {
				Email string `json:"email"`
			} `json:"customer"`
			Metadata struct {
				CompanyID string `json:"company_id"`
				Plan      string `json:"plan"`
			} `json:"metadata"`
		} `json:"data"`
	}

	if err := c.BodyParser(&event); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// We only care about successful charges
	if event.Event != "charge.success" || event.Data.Status != "success" {
		return c.SendStatus(fiber.StatusOK)
	}

	companyIDStr := event.Data.Metadata.CompanyID
	if companyIDStr == "" {
		logger.Error().Str("reference", event.Data.Reference).Msg("Webhook missing company_id in metadata")
		return c.SendStatus(fiber.StatusBadRequest)
	}

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		logger.Error().Str("company_id", companyIDStr).Msg("Invalid company_id in webhook metadata")
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Prevent payment reassignment attacks by verifying customer email matches company admin email
	company, err := h.configUC.GetCompanyByID(c.Context(), companyID)
	if err != nil || company.AdminEmail != event.Data.Customer.Email {
		logger.Error().
			Str("reference", event.Data.Reference).
			Str("webhook_email", event.Data.Customer.Email).
			Str("company_email", company.AdminEmail).
			Msg("Webhook email mismatch or company not found")
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	planType := event.Data.Metadata.Plan

	// Resolve expected price in kobo — DB first, static fallback
	var expectedPriceKobo int
	dbPlan, dbErr := h.configUC.GetPlanByID(c.Context(), planType)
	if dbErr == nil {
		expectedPriceKobo = int(dbPlan.BasePrice)
	} else {
		staticPlan, staticErr := billing.GetPlanByID(planType)
		if staticErr != nil {
			staticPlan = billing.PlanPro
		}
		expectedPriceKobo = staticPlan.Price
	}

	// Compare kobo to kobo — no unit conversion needed!
	paidKobo := event.Data.Amount
	if paidKobo < expectedPriceKobo {
		logger.Error().
			Int("paid_kobo", paidKobo).
			Int("expected_kobo", expectedPriceKobo).
			Str("company", companyIDStr).
			Msg("Insufficient payment amount for plan")
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Record the payment to guarantee idempotency
	// Store the human-readable Naira amount in the payments table
	amountNaira := float64(paidKobo) / 100.0
	id, err := h.configUC.RecordPayment(c.Context(), companyID, event.Data.Reference, amountNaira, "success")
	if err != nil {
		logger.Error().Err(err).Str("reference", event.Data.Reference).Msg("Failed to record payment")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if id == 0 {
		logger.Info().Str("reference", event.Data.Reference).Msg("Duplicate payment webhook ignored")
		return c.SendStatus(fiber.StatusOK)
	}

	err = h.configUC.UpdateCompanySubscriptionStatus(c.Context(), companyID, "active", planType)
	if err != nil {
		logger.Error().Err(err).Str("company_id", companyIDStr).Msg("Failed to update subscription status on payment")
	} else {
		logger.Info().Str("company_id", companyIDStr).Str("plan", planType).Msg("Company subscription activated via Paystack")
	}

	return c.SendStatus(fiber.StatusOK)
}

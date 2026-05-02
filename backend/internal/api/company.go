package api

import (
	"strings"
	"time"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/payment"
	"webtracker-bot/internal/shipment"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/google/uuid"
)

type CompanyHandler struct {
	cfg      *config.Config
	configUC models.ConfigUsecase
	bots     models.BotProvider
	paystack *payment.PaystackService
}

func NewCompanyHandler(cfg *config.Config, configUC models.ConfigUsecase, bots models.BotProvider) *CompanyHandler {
	return &CompanyHandler{
		cfg:      cfg,
		configUC: configUC,
		bots:     bots,
		paystack: payment.NewPaystackService(cfg.PaystackSecretKey),
	}
}

func (h *CompanyHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/company")

	// Limit pairing requests to prevent WhatsApp spam
	pairLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return getCompanyID(c).String() // rate limit per company rather than IP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many pairing attempts. Please wait a minute.",
			})
		},
	})

	api.Post("/activate", h.activateBot)
	api.Post("/deactivate", h.deactivateBot)
	api.Post("/pair", pairLimiter, h.pairBot)
	api.Post("/qr", pairLimiter, h.getQR)
	api.Post("/logout", h.logoutBot)
	api.Delete("/delete", h.deleteCompany)

	api.Post("/subscribe", h.subscribe)
	api.Get("/subscription-status", h.getSubscriptionStatus)

	// Webhooks
	app.Post("/api/webhooks/paystack", h.paystackWebhook)
}

func (h *CompanyHandler) checkSubscription(ctx *fiber.Ctx, companyID uuid.UUID) error {
	// Super admin bypasses all billing checks
	if shipment.IsSuperAdmin(h.cfg, companyID) {
		return nil
	}

	company, err := h.configUC.GetCompanyByID(ctx.Context(), companyID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to verify subscription status")
	}

	if company.SubscriptionStatus.String != "active" && company.SubscriptionStatus.String != "trialing" {
		return fiber.NewError(fiber.StatusPaymentRequired, "Subscription is inactive. Please renew your subscription to use the tracking bot.")
	}

	if company.SubscriptionExpiry.Valid && company.SubscriptionExpiry.Time.Before(time.Now()) {
		return fiber.NewError(fiber.StatusPaymentRequired, "Subscription has expired. Please renew to continue using the tracking bot.")
	}

	return nil
}

func (h *CompanyHandler) activateBot(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	if err := h.checkSubscription(c, companyID); err != nil {
		return err
	}

	if err := h.bots.ActivateBot(c.Context(), companyID); err != nil {
		logger.Error().Err(err).Str("company", companyID.String()).Msg("Failed to activate bot")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Bot activated successfully"})
}

func (h *CompanyHandler) deactivateBot(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	if err := h.bots.DeactivateBot(companyID); err != nil {
		logger.Error().Err(err).Str("company", companyID.String()).Msg("Failed to deactivate bot")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Bot deactivated successfully"})
}

func (h *CompanyHandler) logoutBot(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	if err := h.bots.LogoutBot(companyID); err != nil {
		// If the bot is already gone, that's fine — the intent was to disconnect
		if strings.Contains(err.Error(), "bot not found") {
			logger.Info().Str("company", companyID.String()).Msg("Bot already disconnected")
			return c.JSON(fiber.Map{"success": true, "message": "Bot already disconnected"})
		}
		logger.Error().Err(err).Str("company", companyID.String()).Msg("Failed to logout bot")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Bot logged out successfully"})
}

func (h *CompanyHandler) deleteCompany(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	// Try to logout the bot first (best-effort, ignore errors)
	_ = h.bots.LogoutBot(companyID)

	// Delete all company data
	if err := h.configUC.DeleteCompany(c.Context(), companyID); err != nil {
		logger.Error().Err(err).Str("company", companyID.String()).Msg("Failed to delete company")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete account. Please try again."})
	}

	logger.Info().Str("company_id", companyID.String()).Msg("Company account permanently deleted")
	return c.JSON(fiber.Map{"success": true, "message": "Account permanently deleted"})
}

func (h *CompanyHandler) pairBot(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	if err := h.checkSubscription(c, companyID); err != nil {
		return err
	}

	var req struct {
		Phone string `json:"phone"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "phone is required"})
	}

	code, err := h.bots.GeneratePairingCode(c.Context(), companyID, req.Phone)
	if err != nil {
		logger.Error().Err(err).Str("company", companyID.String()).Msg("Failed to generate pairing code")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "code": code})
}

func (h *CompanyHandler) getQR(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	if err := h.checkSubscription(c, companyID); err != nil {
		return err
	}

	code, err := h.bots.GetQR(c.Context(), companyID)
	if err != nil {
		logger.Error().Err(err).Str("company", companyID.String()).Msg("Failed to generate QR code")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "code": code})
}

// subscribe initializes a Paystack transaction for the company
func (h *CompanyHandler) subscribe(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
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

	dbPlan, dbErr := h.configUC.GetPlanByID(c.Context(), req.Plan)
	var planPrice int
	var planID string

	if dbErr == nil {
		planPrice = int(dbPlan.BasePrice)
		planID = dbPlan.ID
	} else {
		// Fallback to static if DB not seeded
		plan, err := payment.GetPlanByID(req.Plan)
		if err != nil {
			plan = payment.PlanPro // Default
		}
		planPrice = plan.Price
		planID = plan.ID
	}

	metadata := map[string]interface{}{
		"company_id": companyID.String(),
		"plan":       planID,
	}

	authURL, reference, err := h.paystack.InitializeTransaction(company.AdminEmail, planPrice, req.CallbackURL, metadata)
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

func (h *CompanyHandler) getSubscriptionStatus(c *fiber.Ctx) error {
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

// paystackWebhook handles inbound webhook events from Paystack
func (h *CompanyHandler) paystackWebhook(c *fiber.Ctx) error {
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
			Amount    int    `json:"amount"`
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
	if event.Event == "charge.success" && event.Data.Status == "success" {
		companyIDStr := event.Data.Metadata.CompanyID
		if companyIDStr != "" {
			if companyID, err := uuid.Parse(companyIDStr); err == nil {
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

				// Record the payment to guarantee idempotency
				amountInNaira := float64(event.Data.Amount) / 100.0
				id, err := h.configUC.RecordPayment(c.Context(), companyID, event.Data.Reference, amountInNaira, "success")
				if err != nil {
					logger.Error().Err(err).Str("reference", event.Data.Reference).Msg("Failed to record payment")
					return c.SendStatus(fiber.StatusInternalServerError)
				}

				if id == 0 {
					logger.Info().Str("reference", event.Data.Reference).Msg("Duplicate payment webhook ignored")
					return c.SendStatus(fiber.StatusOK)
				}

				planType := event.Data.Metadata.Plan
				err = h.configUC.UpdateCompanySubscriptionStatus(c.Context(), companyID, "active", planType)
				if err != nil {
					logger.Error().Err(err).Str("company_id", companyIDStr).Msg("Failed to update subscription status on payment")
				} else {
					logger.Info().Str("company_id", companyIDStr).Str("plan", planType).Msg("Company subscription activated via Paystack")
				}
			}
		}
	}

	return c.SendStatus(fiber.StatusOK)
}

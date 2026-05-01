package api

import (
	"crypto/rand"
	"encoding/hex"
	"time"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/payment"
	"webtracker-bot/internal/whatsapp"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CompanyHandler struct {
	cfg      *config.Config
	configUC *config.Usecase
	bots     whatsapp.BotProvider
	paystack *payment.PaystackService
}

func NewCompanyHandler(cfg *config.Config, configUC *config.Usecase, bots whatsapp.BotProvider) *CompanyHandler {
	return &CompanyHandler{
		cfg:      cfg,
		configUC: configUC,
		bots:     bots,
		paystack: payment.NewPaystackService(cfg.PaystackSecretKey),
	}
}

func (h *CompanyHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/company")

	// Multi-tenant: frontend passes X-Company-ID header

	api.Post("/activate", h.activateBot)
	api.Post("/deactivate", h.deactivateBot)
	api.Post("/pair", h.pairBot)
	api.Post("/qr", h.getQR)
	api.Post("/logout", h.logoutBot)
	api.Delete("/delete", h.deleteCompany)

	api.Get("/setup/:token", h.getCompanyBySetupToken)
	api.Post("/onboard", h.onboardCompany)
	api.Post("/resend-setup-link", h.resendSetupLink)
	api.Post("/subscribe", h.subscribe)

	// Webhooks
	app.Post("/api/webhooks/paystack", h.paystackWebhook)
}

func (h *CompanyHandler) getCompanyBySetupToken(c *fiber.Ctx) error {
	token := c.Params("token")
	company, err := h.configUC.GetCompanyBySetupToken(c.Context(), token)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invalid or expired setup token"})
	}

	return c.JSON(fiber.Map{
		"id":          company.ID,
		"name":        company.Name.String,
		"auth_status": company.AuthStatus.String,
	})
}

func (h *CompanyHandler) checkSubscription(ctx *fiber.Ctx, companyID uuid.UUID) error {
	// Super admin bypasses all billing checks
	if IsSuperAdmin(h.cfg, companyID) {
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


func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// onboardCompany creates a new company and sends a magic setup link email
func (h *CompanyHandler) onboardCompany(c *fiber.Ctx) error {
	var req struct {
		Name       string `json:"name"`
		AdminEmail string `json:"admin_email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Name == "" || req.AdminEmail == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name and admin_email are required"})
	}

	token := generateToken()

	company, err := h.configUC.CreateCompany(c.Context(), req.Name, req.AdminEmail, token)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create company")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Send magic link email asynchronously
	go notif.SendSetupLinkEmail(h.cfg, req.AdminEmail, req.Name, token)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":    true,
		"company_id": company.ID,
		"message":    "Company created. Setup link sent to " + req.AdminEmail,
	})
}

// resendSetupLink regenerates a setup token and resends the magic link email
func (h *CompanyHandler) resendSetupLink(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	company, err := h.configUC.GetCompanyByID(c.Context(), companyID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Company not found"})
	}

	newToken := generateToken()
	if err := h.configUC.RegenerateSetupToken(c.Context(), companyID, newToken); err != nil {
		logger.Error().Err(err).Msg("Failed to regenerate setup token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to regenerate token"})
	}

	go notif.SendSetupLinkEmail(h.cfg, company.AdminEmail, company.Name.String, newToken)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Setup link resent to " + company.AdminEmail,
	})
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

	var amount int
	switch req.Plan {
	case "starter":
		amount = 1490000 // ₦14,900 in kobo
	case "enterprise", "scale":
		amount = 22500000 // ₦225,000 in kobo
	default:
		// Default to Pro
		amount = 5990000 // ₦59,900 in kobo
		req.Plan = "pro"
	}

	metadata := map[string]interface{}{
		"company_id": companyID.String(),
		"plan":       req.Plan,
	}

	authURL, err := h.paystack.InitializeTransaction(company.AdminEmail, amount, req.CallbackURL, metadata)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize Paystack transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start payment"})
	}

	return c.JSON(fiber.Map{
		"success":           true,
		"authorization_url": authURL,
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
			Status   string `json:"status"`
			Metadata struct {
				CompanyID string `json:"company_id"`
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
				err := h.configUC.UpdateCompanySubscriptionStatus(c.Context(), companyID, "active")
				if err != nil {
					logger.Error().Err(err).Str("company_id", companyIDStr).Msg("Failed to update subscription status on payment")
				} else {
					logger.Info().Str("company_id", companyIDStr).Msg("Company subscription activated via Paystack")
				}
			}
		}
	}

	return c.SendStatus(fiber.StatusOK)
}


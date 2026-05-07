package api

import (
	"strings"
	"time"
	"webtracker-bot/internal/billing"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/google/uuid"
)

type CompanyHandler struct {
	cfg      *config.Config
	configUC models.ConfigUsecase
	bots     models.BotProvider
}

func NewCompanyHandler(cfg *config.Config, configUC models.ConfigUsecase, bots models.BotProvider) *CompanyHandler {
	return &CompanyHandler{
		cfg:      cfg,
		configUC: configUC,
		bots:     bots,
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
}

func (h *CompanyHandler) checkSubscription(ctx *fiber.Ctx, companyID uuid.UUID) error {
	// Super admin bypasses all billing checks
	if billing.IsSuperAdmin(h.cfg, companyID) {
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

	// Purge the bot and its paired device from the store
	_ = h.bots.PurgeBot(companyID)

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



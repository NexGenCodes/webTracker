package api

import (
	"errors"
	"strings"
	"time"
	"webtracker-bot/internal/auth"
	"webtracker-bot/internal/billing"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/whatsapp"

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

	api.Put("/settings", h.updateSettings)
	api.Post("/activate", h.activateBot)
	api.Post("/deactivate", h.deactivateBot)
	api.Post("/pair", pairLimiter, h.pairBot)
	api.Post("/qr", pairLimiter, h.getQR)
	api.Post("/logout", h.logoutBot)
	api.Delete("/delete", h.deleteCompany)
}

func (h *CompanyHandler) checkSubscription(ctx *fiber.Ctx, companyID uuid.UUID) error {
	company, err := h.configUC.GetCompanyByID(ctx.Context(), companyID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to verify subscription status")
	}

	// Super admin bypasses all billing checks (RBAC: check JWT role, not email)
	if claims, ok := ctx.Locals("user").(*auth.JWTClaims); ok && billing.IsSuperAdminRole(claims.Role) {
		return nil
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

	// Purge the bot and its paired device from the store (best-effort)
	if err := h.bots.PurgeBot(companyID); err != nil {
		logger.Warn().Err(err).Str("company_id", companyID.String()).Msg("Company: purge bot before delete had non-critical error")
	}

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
		if errors.Is(err, whatsapp.ErrAlreadyPaired) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "already_connected"})
		}
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
		if errors.Is(err, whatsapp.ErrAlreadyPaired) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "already_connected"})
		}
		logger.Error().Err(err).Str("company", companyID.String()).Msg("Failed to generate QR code")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "code": code})
}


type UpdateSettingsRequest struct {
	Name           string `json:"name"`
	AdminEmail     string `json:"admin_email"`
	LogoUrl        string `json:"logo_url"`
	BrandColor     string `json:"brand_color"`
	TrackingPrefix string `json:"tracking_prefix"`
}

func (h *CompanyHandler) updateSettings(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	var req UpdateSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}
	if len(req.TrackingPrefix) > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tracking_prefix must be 5 characters or fewer"})
	}
	if req.BrandColor != "" && !strings.HasPrefix(req.BrandColor, "#") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "brand_color must be a hex code (e.g. #6366f1)"})
	}

	if err := h.configUC.UpdateCompanySettings(c.Context(), companyID, req.Name, req.AdminEmail, req.LogoUrl, req.BrandColor, req.TrackingPrefix); err != nil {
		logger.Error().Err(err).Str("company", companyID.String()).Msg("Failed to update company settings")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update settings"})
	}

	// Emit audit log — best-effort, non-blocking
	actorEmail := ""
	if user, ok := c.Locals("user").(*auth.JWTClaims); ok && user != nil {
		actorEmail = user.Email
	}
	if err := h.configUC.LogAudit(c.Context(), actorEmail, "update_settings", companyID, map[string]interface{}{
		"name":            req.Name,
		"admin_email":     req.AdminEmail,
		"tracking_prefix": req.TrackingPrefix,
	}); err != nil {
		logger.Error().Err(err).Str("company_id", companyID.String()).Msg("Failed to log audit for update_settings")
	}

	return c.JSON(fiber.Map{"success": true})
}

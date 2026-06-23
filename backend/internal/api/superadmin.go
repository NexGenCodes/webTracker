package api

import (
	"strconv"
	"time"
	"webtracker-bot/internal/auth"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SuperAdminHandler struct {
	cfg      *config.Config
	configUC models.ConfigUsecase
	bots     models.BotProvider
}

func NewSuperAdminHandler(cfg *config.Config, configUC models.ConfigUsecase, bots models.BotProvider) *SuperAdminHandler {
	return &SuperAdminHandler{
		cfg:      cfg,
		configUC: configUC,
		bots:     bots,
	}
}

func (h *SuperAdminHandler) RegisterRoutes(app *fiber.App) {
	admin := app.Group("/api/super-admin", auth.SuperAdminMiddleware())

	admin.Get("/companies", h.listCompanies)
	admin.Delete("/companies/:id", h.deleteCompany)
	admin.Put("/companies/:id/plan", h.updatePlan)
	admin.Put("/companies/:id/subscription", h.updateSubscription)
	admin.Post("/companies/:id/disconnect-bot", h.disconnectBot)
	admin.Get("/analytics", h.getAnalytics)
	admin.Get("/audit-log", h.getAuditLogs)
}

func (h *SuperAdminHandler) listCompanies(c *fiber.Ctx) error {
	companies, err := h.configUC.GetAllCompanyDetails(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch companies"})
	}
	return c.JSON(companies)
}

func (h *SuperAdminHandler) deleteCompany(c *fiber.Ctx) error {
	idStr := c.Params("id")
	companyID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid company ID"})
	}

	actor := c.Locals("user").(*auth.JWTClaims).Email

	// Purge bot (best-effort — proceed even if bot is already gone)
	if err := h.bots.PurgeBot(companyID); err != nil {
		logger.Warn().Err(err).Str("company_id", idStr).Msg("Super admin: purge bot before delete had non-critical error")
	}

	// Delete data
	if err := h.configUC.DeleteCompany(c.Context(), companyID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Audit log — recorded AFTER successful deletion
	if err := h.configUC.LogAudit(c.Context(), actor, "delete_company", companyID, map[string]interface{}{"company_id": idStr}); err != nil {
		logger.Error().Err(err).Msg("Super admin: failed to log audit for delete_company")
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *SuperAdminHandler) updatePlan(c *fiber.Ctx) error {
	idStr := c.Params("id")
	companyID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid company ID"})
	}

	var req struct {
		PlanType string `json:"plan_type"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.configUC.UpdateCompanyPlan(c.Context(), companyID, req.PlanType); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Audit log — recorded AFTER successful update
	actor := c.Locals("user").(*auth.JWTClaims).Email
	if err := h.configUC.LogAudit(c.Context(), actor, "change_plan", companyID, map[string]interface{}{"new_plan": req.PlanType}); err != nil {
		logger.Error().Err(err).Msg("Super admin: failed to log audit for change_plan")
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *SuperAdminHandler) updateSubscription(c *fiber.Ctx) error {
	idStr := c.Params("id")
	companyID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid company ID"})
	}

	var req struct {
		Status string    `json:"status"`
		Expiry time.Time `json:"expiry"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.configUC.UpdateCompanySubscription(c.Context(), companyID, req.Status, req.Expiry); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Audit log — recorded AFTER successful update
	actor := c.Locals("user").(*auth.JWTClaims).Email
	if err := h.configUC.LogAudit(c.Context(), actor, "update_subscription", companyID, map[string]interface{}{
		"status": req.Status,
		"expiry": req.Expiry,
	}); err != nil {
		logger.Error().Err(err).Msg("Super admin: failed to log audit for update_subscription")
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *SuperAdminHandler) disconnectBot(c *fiber.Ctx) error {
	idStr := c.Params("id")
	companyID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid company ID"})
	}

	actor := c.Locals("user").(*auth.JWTClaims).Email
	if err := h.configUC.LogAudit(c.Context(), actor, "force_disconnect_bot", companyID, nil); err != nil {
		logger.Error().Err(err).Msg("Super admin: failed to log audit for force_disconnect_bot")
	}

	if err := h.bots.LogoutBot(companyID); err != nil {
		logger.Error().Err(err).Str("company", companyID.String()).Msg("Super admin failed to force logout bot")
	}

	return c.JSON(fiber.Map{"success": true, "message": "Bot disconnect command sent"})
}

func (h *SuperAdminHandler) getAnalytics(c *fiber.Ctx) error {
	stats, err := h.configUC.GetPlatformAnalytics(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(stats)
}

func (h *SuperAdminHandler) getAuditLogs(c *fiber.Ctx) error {
	limit, err := strconv.Atoi(c.Query("limit", "50"))
	if err != nil {
		limit = 50
	}
	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil {
		offset = 0
	}

	logs, err := h.configUC.GetAuditLogs(c.Context(), int32(limit), int32(offset))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(logs)
}

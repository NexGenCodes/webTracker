package auth

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"webtracker-bot/internal/logger"
)

type Handler struct {
	service    *Service
	validate   *validator.Validate
	isSecure   bool
	sameSite   string
}

func NewHandler(service *Service) *Handler {
	isSecure := !strings.HasPrefix(service.cfg.FrontendURL, "http://localhost")
	sameSite := "Lax"
	if isSecure {
		sameSite = "None"
	}
	return &Handler{
		service:  service,
		validate: validator.New(),
		isSecure: isSecure,
		sameSite: sameSite,
	}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	group := app.Group("/api/auth")

	group.Post("/register-intent", h.registerIntent)
	group.Post("/verify-otp", h.verifyOTP)
	group.Post("/login", h.login)
	group.Post("/logout", h.logout)
	group.Post("/forgot-password", h.forgotPassword)
	group.Post("/reset-password", h.resetPassword)
	
	// Protected routes
	group.Get("/me", JWTAuth(h.service.cfg.JWTSecret), h.me)
	group.Post("/setup", JWTAuth(h.service.cfg.JWTSecret), h.setupCompany)
}

func (h *Handler) forgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	resetToken, err := h.service.InitiatePasswordReset(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "reset_token",
		Value:    resetToken,
		Expires:  time.Now().Add(15 * time.Minute),
		HTTPOnly: true,
		Secure:   h.isSecure,
		SameSite: h.sameSite,
		Path:     "/",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Reset code sent", "reset_token": resetToken})
}

func (h *Handler) resetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	resetToken := c.Cookies("reset_token")
	if resetToken == "" {
		resetToken = c.Get("X-Reset-Token")
	}

	if resetToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Reset session expired or missing"})
	}

	err := h.service.CompletePasswordReset(c.Context(), req, resetToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	// Clear reset token
	c.Cookie(&fiber.Cookie{Name: "reset_token", Value: "", Expires: time.Now().Add(-1 * time.Hour), HTTPOnly: true, Path: "/"})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Password reset successful"})
}


func (h *Handler) me(c *fiber.Ctx) error {
	user := c.Locals("user").(*JWTClaims)
	return c.JSON(fiber.Map{
		"company_id":   user.CompanyID,
		"company_name": user.CompanyName,
		"email":        user.Email,
		"plan_type":    user.PlanType,
		"auth_status":  user.AuthStatus,
	})
}

func (h *Handler) registerIntent(c *fiber.Ctx) error {
	var req RegisterIntentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	otpToken, err := h.service.GenerateOTP(c.Context(), req.CompanyName, req.Email, req.Password)
	if err != nil {
		logger.Error().Err(err).Msg("Registration intent failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Set temporary OTP cookie
	c.Cookie(&fiber.Cookie{
		Name:     "otp_token",
		Value:    otpToken,
		Expires:  time.Now().Add(10 * time.Minute),
		HTTPOnly: true,
		Secure:   h.isSecure,
		SameSite: h.sameSite,
		Path:     "/",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "OTP sent to email", "otp_token": otpToken})
}

func (h *Handler) verifyOTP(c *fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	otpToken := c.Cookies("otp_token")
	if otpToken == "" {
		otpToken = c.Get("X-OTP-Token")
	}
	
	if otpToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "OTP session expired or missing"})
	}

	resp, sessionToken, err := h.service.VerifyOTP(c.Context(), req.OTP, otpToken)
	if err != nil {
		logger.Error().Err(err).Msg("OTP Verification failed")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	// Clear OTP token and set Session token
	c.Cookie(&fiber.Cookie{Name: "otp_token", Value: "", Expires: time.Now().Add(-1 * time.Hour), HTTPOnly: true, Path: "/"})
	h.setJWTCookie(c, sessionToken)
	resp.Token = sessionToken
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *Handler) setupCompany(c *fiber.Ctx) error {
	user := c.Locals("user").(*JWTClaims)
	
	var req SetupCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	resp, sessionToken, err := h.service.SetupCompany(c.Context(), user.CompanyID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.setJWTCookie(c, sessionToken)
	resp.Token = sessionToken
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *Handler) login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	resp, tokenString, err := h.service.Login(c.Context(), req)
	if err != nil {
		logger.Warn().Err(err).Msg("Login failed")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	h.setJWTCookie(c, tokenString)
	resp.Token = tokenString
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *Handler) logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   h.isSecure,
		SameSite: h.sameSite,
		Path:     "/",
	})
	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) setJWTCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		Secure:   h.isSecure,
		SameSite: h.sameSite,
		Path:     "/",
	})
}

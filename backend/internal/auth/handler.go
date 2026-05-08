package auth

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

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
	// Create a rate limiter: Max 10 requests per minute per IP for auth routes
	authLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many attempts, please try again later.",
			})
		},
	})

	group := app.Group("/api/auth", authLimiter)

	group.Post("/register-intent", h.registerIntent)
	group.Post("/verify-otp", h.verifyOTP)
	group.Post("/login", h.login)
	group.Post("/admin-login", h.adminLogin)
	group.Post("/logout", h.logout)
	group.Post("/forgot-password", h.forgotPassword)
	group.Post("/reset-password", h.resetPassword)
	
	// Protected routes (JWT validation is handled by global middleware in server.go)
	group.Get("/me", h.me)
	group.Post("/setup", h.setupCompany)
}

func (h *Handler) forgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Always return 200 regardless of outcome to prevent email enumeration.
	// Internal errors are logged but never exposed to the client.
	if err := h.service.InitiatePasswordReset(c.Context(), req.Email); err != nil {
		logger.Error().Err(err).Str("email", req.Email).Msg("Password reset initiation failed internally")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "If that email exists, a reset code has been sent."})
}

func (h *Handler) resetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	err := h.service.CompletePasswordReset(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Password reset successful"})
}


func (h *Handler) me(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*JWTClaims)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
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

	err := h.service.GenerateOTP(c.Context(), req.CompanyName, req.Email, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		logger.Error().Err(err).Msg("Registration intent failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process registration"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "OTP sent to email"})
}

func (h *Handler) verifyOTP(c *fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	otpCode := strings.TrimSpace(req.OTP)
	resp, sessionToken, err := h.service.VerifyOTP(c.Context(), req.Email, otpCode)
	if err != nil {
		logger.Error().Err(err).Msg("OTP Verification failed")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	// Set Session token
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

func (h *Handler) adminLogin(c *fiber.Ctx) error {
	var req AdminLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	tokenString, err := h.service.AdminLogin(c.Context(), req)
	if err != nil {
		logger.Warn().Err(err).Msg("Admin Login failed")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	h.setJWTCookie(c, tokenString)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": tokenString})
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

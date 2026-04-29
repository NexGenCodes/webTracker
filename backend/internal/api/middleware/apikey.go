package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// APIKeyAuth returns a Fiber middleware that validates the X-API-Key header
// against the provided secret. Requests without a valid key receive 401.
func APIKeyAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Allow health checks, public tracking, and auth routes without API key
		path := c.Path()
		if path == "/health" || strings.HasPrefix(path, "/api/track/") || strings.HasPrefix(path, "/api/auth/") {
			return c.Next()
		}

		key := c.Get("X-API-Key")
		if key == "" || key != secret {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: invalid or missing API key",
			})
		}
		return c.Next()
	}
}

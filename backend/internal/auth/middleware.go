package auth

import (
	"strings"
	"os"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth returns a middleware that validates a JWT token from cookies or Authorization header
func JWTAuth(publicKeyPath string) fiber.Handler {
	// Read public key once (or lazily)
	var publicKey interface{}
	if keyBytes, err := os.ReadFile(publicKeyPath); err == nil {
		if key, err := jwt.ParseECPublicKeyFromPEM(keyBytes); err == nil {
			publicKey = key
		}
	}

	return func(c *fiber.Ctx) error {
		path := c.Path()
		// Allow health checks and specific auth/webhook routes without JWT
		if path == "/health" ||
			path == "/api/auth/register-intent" ||
			path == "/api/auth/verify-otp" ||
			path == "/api/auth/login" ||
			path == "/api/auth/logout" ||
			path == "/api/auth/forgot-password" ||
			path == "/api/auth/reset-password" ||
			path == "/api/billing/plans" ||
			strings.HasPrefix(path, "/api/webhooks/") {
			return c.Next()
		}

		// First try cookie
		tokenString := c.Cookies("jwt")

		// Fallback to Bearer token
		if tokenString == "" {
			authHeader := c.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authentication token",
			})
		}

		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			if publicKey == nil {
				// Try reloading just in case
				if keyBytes, err := os.ReadFile(publicKeyPath); err == nil {
					if key, err := jwt.ParseECPublicKeyFromPEM(keyBytes); err == nil {
						publicKey = key
					}
				}
			}
			if publicKey == nil {
				return nil, fmt.Errorf("public key not loaded")
			}
			return publicKey, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		// Set the claims in the context for downstream handlers
		c.Locals("user", claims)

		return c.Next()
	}
}

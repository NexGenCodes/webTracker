package auth

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth returns a middleware that validates a JWT token from cookies or Authorization header
func JWTAuth(publicKeyPath string) fiber.Handler {
	var publicKey interface{}
	var mu sync.RWMutex
	var lastAttempt time.Time

	loadKey := func() interface{} {
		mu.RLock()
		if publicKey != nil {
			defer mu.RUnlock()
			return publicKey
		}
		mu.RUnlock()

		mu.Lock()
		defer mu.Unlock()
		if publicKey != nil {
			return publicKey
		}

		// Throttle disk reads to once every 10 seconds if file is missing
		if time.Since(lastAttempt) < 10*time.Second {
			return nil
		}
		lastAttempt = time.Now()

		if keyBytes, err := os.ReadFile(publicKeyPath); err == nil {
			if key, err := jwt.ParseECPublicKeyFromPEM(keyBytes); err == nil {
				publicKey = key
			}
		}
		return publicKey
	}

	// Try loading initially
	loadKey()

	return func(c *fiber.Ctx) error {
		path := c.Path()
		// Allow health checks and specific auth/webhook routes without JWT
		if path == "/health" ||
			path == "/api/auth/register-intent" ||
			path == "/api/auth/verify-otp" ||
			path == "/api/auth/login" ||
			path == "/api/auth/admin-login" ||
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
			
			key := loadKey()
			if key == nil {
				return nil, fmt.Errorf("public key not loaded")
			}
			return key, nil
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

// SuperAdminMiddleware restricts access to users with the "super_admin" role
func SuperAdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*JWTClaims)
		if !ok || user.AppRole != "super_admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied: super admin privileges required",
			})
		}
		return c.Next()
	}
}

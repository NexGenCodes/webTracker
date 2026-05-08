package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig configures the Redis-backed rate limiter.
type RateLimitConfig struct {
	// Max requests allowed per window.
	Max int
	// Window duration (e.g. 1 * time.Minute).
	Window time.Duration
	// KeyFunc derives the rate-limit key from the request (default: IP).
	KeyFunc func(c *fiber.Ctx) string
	// LimitReached is the handler called when the limit is exceeded.
	LimitReached fiber.Handler
}

// NewRedisRateLimiter returns a Fiber middleware backed by Redis for distributed rate limiting.
func NewRedisRateLimiter(cfg RateLimitConfig) fiber.Handler {
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(c *fiber.Ctx) string { return c.IP() }
	}
	if cfg.LimitReached == nil {
		cfg.LimitReached = func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "too many requests, please try again later",
			})
		}
	}

	return func(c *fiber.Ctx) error {
		if RedisClient == nil {
			return c.Next()
		}

		key := fmt.Sprintf("rl:%s", cfg.KeyFunc(c))
		ctx := c.Context()

		// Use a pipeline for atomic Increment + Expire
		pipe := RedisClient.Pipeline()
		incr := pipe.Incr(ctx, key)
		// ExpireNX: only set expiry on first increment to preserve the rate window boundary
		pipe.ExpireNX(ctx, key, cfg.Window)
		
		if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
			// Fail-open on Redis error
			return c.Next()
		}

		count := incr.Val()
		if count > int64(cfg.Max) {
			return cfg.LimitReached(c)
		}

		remaining := int64(cfg.Max) - count
		if remaining < 0 {
			remaining = 0
		}
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Max))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		return c.Next()
	}
}

// LoginFailLimiter handles brute-force protection using Redis counters.
type LoginFailLimiter struct {
	max    int
	window time.Duration
}

func NewLoginFailLimiter(max int, window time.Duration) *LoginFailLimiter {
	return &LoginFailLimiter{max: max, window: window}
}

func (l *LoginFailLimiter) key(email string) string {
	return fmt.Sprintf("login_fail:%s", email)
}

func (l *LoginFailLimiter) Increment(ctx context.Context, email string) (int64, error) {
	if RedisClient == nil {
		return 0, nil
	}
	pipe := RedisClient.Pipeline()
	incr := pipe.Incr(ctx, l.key(email))
	pipe.Expire(ctx, l.key(email), l.window)
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return 0, err
	}
	return incr.Val(), nil
}

func (l *LoginFailLimiter) IsBlocked(ctx context.Context, email string) (bool, error) {
	if RedisClient == nil {
		return false, nil
	}
	count, err := RedisClient.Get(ctx, l.key(email)).Int64()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return count >= int64(l.max), nil
}

func (l *LoginFailLimiter) Reset(ctx context.Context, email string) {
	if RedisClient != nil {
		RedisClient.Del(ctx, l.key(email))
	}
}

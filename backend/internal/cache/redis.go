package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"webtracker-bot/internal/logger"
)

var (
	RedisClient *redis.Client
	AsynqClient *asynq.Client
	mu          sync.Mutex
	initialized bool
)

// InitRedis initializes the global Redis and Asynq clients.
// Safe to retry on failure — only succeeds once.
func InitRedis(redisURL string) error {
	mu.Lock()
	defer mu.Unlock()

	if initialized {
		return nil
	}

	// Initialize standard Redis client for caching, OTPs, Rate Limiting
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("invalid REDIS_URL (expected redis://host:port): %w", err)
	}

	client := redis.NewClient(opt)

	// Ping test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	RedisClient = client

	// Initialize Asynq Client for Task Queues
	redisOpt := asynq.RedisClientOpt{Addr: opt.Addr, Password: opt.Password, DB: opt.DB}
	AsynqClient = asynq.NewClient(redisOpt)

	initialized = true
	logger.Info().Str("addr", opt.Addr).Msg("Connected to Redis and initialized Asynq client")

	return nil
}

// Close gracefully closes the Redis connections
func Close() {
	mu.Lock()
	defer mu.Unlock()

	if !initialized {
		return
	}

	if AsynqClient != nil {
		AsynqClient.Close()
		AsynqClient = nil
	}
	if RedisClient != nil {
		RedisClient.Close()
		RedisClient = nil
	}
	initialized = false
}

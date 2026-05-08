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
	once        sync.Once
)

// InitRedis initializes the global Redis and Asynq clients
func InitRedis(redisURL string) error {
	var initErr error

	once.Do(func() {
		// Initialize standard Redis client for caching, OTPs, Rate Limiting
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			initErr = fmt.Errorf("invalid REDIS_URL: %w", err)
			return
		}

		RedisClient = redis.NewClient(opt)

		// Ping test
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := RedisClient.Ping(ctx).Err(); err != nil {
			initErr = fmt.Errorf("failed to connect to Redis: %w", err)
			return
		}

		// Initialize Asynq Client for Task Queues
		redisOpt := asynq.RedisClientOpt{Addr: opt.Addr, Password: opt.Password, DB: opt.DB}
		AsynqClient = asynq.NewClient(redisOpt)

		logger.Info().Str("addr", opt.Addr).Msg("Connected to Redis and initialized Asynq client")
	})

	return initErr
}

// Close gracefully closes the Redis connections
func Close() {
	if AsynqClient != nil {
		AsynqClient.Close()
	}
	if RedisClient != nil {
		RedisClient.Close()
	}
}

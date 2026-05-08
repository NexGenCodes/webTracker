package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Get fetches a cached value and unmarshals it into dest.
// Returns (false, nil) on cache miss, (false, err) on error.
func Get[T any](ctx context.Context, key string, dest *T) (bool, error) {
	if RedisClient == nil {
		return false, nil
	}
	data, err := RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return false, err
	}
	return true, nil
}

// Set marshals value and stores it with the given TTL.
func Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if RedisClient == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache: marshal error for key %q: %w", key, err)
	}
	return RedisClient.Set(ctx, key, data, ttl).Err()
}

// Del removes one or more keys from Redis.
func Del(ctx context.Context, keys ...string) {
	if RedisClient == nil {
		return
	}
	RedisClient.Del(ctx, keys...)
}

// ShipmentKey returns the cache key for a single shipment.
func ShipmentKey(companyID, trackingID string) string {
	return fmt.Sprintf("ship:%s:%s", companyID, trackingID)
}

// CompanyKey returns the cache key for a company record.
func CompanyKey(companyID string) string {
	return fmt.Sprintf("company:%s", companyID)
}

const (
	ShipmentTTL = 30 * time.Second  // Short: status updates must propagate quickly
	CompanyTTL  = 5 * time.Minute   // Longer: company config rarely changes
)

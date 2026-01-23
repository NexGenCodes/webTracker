package utils

import (
	"sync"
	"time"
)

type UserLimit struct {
	sync.Mutex
	Count      int
	LastAccess time.Time
}

var (
	userLimits  sync.Map // Map[string]*UserLimit
	maxRequests = 5      // 5 requests
	window      = 60 * time.Second
)

// Allow checks if a user is permitted to perform an action.
// It is high-performance and uses zero database calls.
func Allow(userID string) (bool, time.Duration) {
	now := time.Now()
	val, ok := userLimits.Load(userID)
	if !ok {
		userLimits.Store(userID, &UserLimit{Count: 1, LastAccess: now})
		return true, 0
	}

	limit := val.(*UserLimit)
	limit.Lock()
	defer limit.Unlock()

	// Reset window if expired
	if now.Sub(limit.LastAccess) > window {
		limit.Count = 1
		limit.LastAccess = now
		return true, 0
	}

	if limit.Count >= maxRequests {
		retryIn := window - now.Sub(limit.LastAccess)
		return false, retryIn
	}

	limit.Count++
	return true, 0
}

// CleanupLimits removes old entries from the map to save RAM.
// Should be called periodically.
func CleanupLimits() {
	now := time.Now()
	userLimits.Range(func(key, value interface{}) bool {
		limit := value.(*UserLimit)
		if now.Sub(limit.LastAccess) > window {
			userLimits.Delete(key)
		}
		return true
	})
}

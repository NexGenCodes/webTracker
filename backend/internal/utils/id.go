package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateShortID creates a random alphanumeric string of specified length.
// It excludes ambiguous characters: I, 1, O, 0.
func GenerateShortID(length int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// GenerateTrackingID creates a formatted ID: PREFIX-123456789
func GenerateTrackingID(prefix string) string {
	if prefix == "" {
		prefix = "AWB"
	}
	// Generate 9 random digits
	digits := ""
	for i := 0; i < 9; i++ {
		digits += fmt.Sprintf("%d", rand.Intn(10))
	}
	return fmt.Sprintf("%s-%s", prefix, digits)
}

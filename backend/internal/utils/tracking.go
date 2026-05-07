package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateTrackingID creates a collision-resistant tracking ID with the given prefix.
func GenerateTrackingID(prefix string) (string, error) {
	if prefix == "" {
		prefix = "AWB"
	}
	n, randErr := rand.Int(rand.Reader, big.NewInt(1000000000))
	if randErr != nil {
		return "", fmt.Errorf("failed to generate random ID: %w", randErr)
	}
	return fmt.Sprintf("%s-%09d", prefix, n.Int64()), nil
}

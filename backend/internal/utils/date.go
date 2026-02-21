package utils

import (
	"strings"
	"time"
)

// ParseNaturalDate converts words like "today", "tomorrow" into a UTC time.
// Returns the time and true if parsed, zero time and false otherwise.
func ParseNaturalDate(input string, now time.Time) (time.Time, bool) {
	input = strings.ToLower(strings.TrimSpace(input))

	switch input {
	case "today":
		return now, true
	case "tomorrow":
		return now.AddDate(0, 0, 1), true
	case "next tomorrow":
		return now.AddDate(0, 0, 2), true
	case "yesterday":
		return now.AddDate(0, 0, -1), true
	}

	// Also handle simple date formats if needed in the future
	return time.Time{}, false
}

package utils

import (
	"strings"
	"time"
)

// ParseNaturalDate converts words like "today", "tomorrow" into a normalized time.
// All natural dates snap to 9:00 AM in the provided timezone to avoid random timestamps.
func ParseNaturalDate(input string, now time.Time, loc *time.Location) (time.Time, bool) {
	input = strings.ToLower(strings.TrimSpace(input))

	switch input {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, loc), true
	case "tomorrow":
		tomorrow := now.AddDate(0, 0, 1)
		return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 0, 0, 0, loc), true
	case "next tomorrow":
		dayAfter := now.AddDate(0, 0, 2)
		return time.Date(dayAfter.Year(), dayAfter.Month(), dayAfter.Day(), 9, 0, 0, 0, loc), true
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 9, 0, 0, 0, loc), true
	}

	return time.Time{}, false
}

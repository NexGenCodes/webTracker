package utils

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Common date formats users type in WhatsApp messages.
var dateFormats = []string{
	"2006-01-02",          // ISO: 2026-05-15
	"2006-01-02 15:04:05", // ISO+time
	"02/01/2006",          // DD/MM/YYYY (international)
	"2/1/2006",            // D/M/YYYY
	"01/02/2006",          // MM/DD/YYYY (US)
	"1/2/2006",            // M/D/YYYY
	"02-01-2006",          // DD-MM-YYYY
	"2-1-2006",            // D-M-YYYY
	"Jan 2, 2006",         // Jan 15, 2026
	"January 2, 2006",     // January 15, 2026
	"2 Jan 2006",          // 15 Jan 2026
	"2 January 2006",      // 15 January 2026
	"Jan 2",               // Jan 15 (current year)
	"January 2",           // January 15 (current year)
	"2 Jan",               // 15 Jan (current year)
	"2 January",           // 15 January (current year)
}

var relDayRe = regexp.MustCompile(`(?i)^in\s+(\d+)\s+days?$`)

// ParseNaturalDate converts natural language and common date formats into a normalized time.
// All natural dates snap to 9:00 AM in the provided timezone to avoid random timestamps.
func ParseNaturalDate(input string, now time.Time, loc *time.Location) (time.Time, bool) {
	input = strings.TrimSpace(input)
	lower := strings.ToLower(input)

	// 1. Natural language keywords
	switch lower {
	case "now":
		return now.In(loc), true
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, loc), true
	case "tomorrow":
		d := now.AddDate(0, 0, 1)
		return time.Date(d.Year(), d.Month(), d.Day(), 9, 0, 0, 0, loc), true
	case "next tomorrow", "day after tomorrow", "pasado mañana":
		d := now.AddDate(0, 0, 2)
		return time.Date(d.Year(), d.Month(), d.Day(), 9, 0, 0, 0, loc), true
	case "yesterday":
		d := now.AddDate(0, 0, -1)
		return time.Date(d.Year(), d.Month(), d.Day(), 9, 0, 0, 0, loc), true
	case "next week":
		d := now.AddDate(0, 0, 7)
		return time.Date(d.Year(), d.Month(), d.Day(), 9, 0, 0, 0, loc), true
	case "next month":
		d := now.AddDate(0, 1, 0)
		return time.Date(d.Year(), d.Month(), d.Day(), 9, 0, 0, 0, loc), true
	}

	// 2. Relative: "in 3 days", "in 5 days"
	if m := relDayRe.FindStringSubmatch(lower); len(m) == 2 {
		days, err := strconv.Atoi(m[1])
		if err == nil && days > 0 && days <= 365 {
			d := now.AddDate(0, 0, days)
			return time.Date(d.Year(), d.Month(), d.Day(), 9, 0, 0, 0, loc), true
		}
	}

	// 3. Try common date formats (use original input to preserve casing for month names)
	for _, fmt := range dateFormats {
		if t, err := time.ParseInLocation(fmt, input, loc); err == nil {
			// If format has no year component, assume current year
			if t.Year() == 0 {
				t = t.AddDate(now.Year(), 0, 0)
			}
			return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, loc), true
		}
	}

	return time.Time{}, false
}

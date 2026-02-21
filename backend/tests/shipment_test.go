package tests

import (
	"testing"
	"time"
	"webtracker-bot/internal/shipment"
)

func TestCalculateSchedule(t *testing.T) {
	// Base Time: 2023-10-25 10:00 UTC
	baseTime := time.Date(2023, 10, 25, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name              string
		nowUTC            time.Time
		originCountry     string
		destCountry       string
		expectedTransitH  int // Expected Hour (in UTC) for Transit Start
		expectedTransitD  int // Expected Day offset for Transit Start
		expectedDeliveryD int // Expected Day offset for Delivery (from Base)
	}{
		{
			name:             "Mid-Day USA (Lagos Admin Gate Daytime -> 11:00 UTC)",
			nowUTC:           baseTime, // 10:00 UTC = 11:00 AM Lagos
			originCountry:    "usa",    // Daytime in Lagos (8am-9pm) -> Add 1 hour buffer
			expectedTransitH: 11,       // 11:00 AM Lagos = 11:00 UTC (Actually 11 AM Lagos = 10 AM UTC in Oct if UTC+1?)
			// Wait, let's calculate: baseTime is 10:00 UTC.
			// Lagos is UTC+1. So 11:00 AM Lagos.
			// Daytime + 1 hour -> 12:00 PM Lagos.
			// 12:00 PM Lagos = 11:00 AM UTC.
			expectedTransitD: 0,
		},
		{
			name:             "Late Night Nigeria (Lagos is UTC+1 -> 23:00 PM Local)",
			nowUTC:           time.Date(2023, 10, 25, 22, 0, 0, 0, time.UTC), // 22:00 UTC = 23:00 Lagos
			originCountry:    "nigeria",                                      // > 9 PM Rule should apply
			expectedTransitH: 7,                                              // Should wait for 08:00 AM Lagos Next Day = 07:00 UTC
			expectedTransitD: 1,                                              // Next day
		},
		{
			name:             "Optimal Time UK (London is UTC+1 -> 14:00 PM Local)",
			nowUTC:           time.Date(2023, 10, 25, 13, 0, 0, 0, time.UTC), // 13:00 UTC = 14:00 London
			originCountry:    "uk",                                           // Normal hours (8am - 9pm)
			expectedTransitH: 14,                                             // Should add 1 Hour -> 14:00 UTC
			expectedTransitD: 0,                                              // Same day
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transit, _, _ := shipment.CalculateSchedule(tt.nowUTC, tt.originCountry, tt.destCountry)

			// Check Transit Time
			if transit.Hour() != tt.expectedTransitH {
				t.Errorf("Transit Hour: got %d, want %d", transit.Hour(), tt.expectedTransitH)
			}

			// Day check (very basic)
			dayDiff := transit.Day() - tt.nowUTC.Day()
			if dayDiff < 0 {
				dayDiff += 30
			}

			if dayDiff != tt.expectedTransitD {
				t.Errorf("Transit Day Offset: got %d, want %d", dayDiff, tt.expectedTransitD)
			}
		})
	}
}

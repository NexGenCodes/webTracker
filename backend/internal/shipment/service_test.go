package shipment

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAlgorithmA_Departure(t *testing.T) {
	calc := &Calculator{}
	adminTZ := "Africa/Lagos" // UTC+1

	tests := []struct {
		name     string
		now      time.Time
		expected string // Lagos local time string
	}{
		{
			name:     "Normal Day Hour (2 PM Lagos)",
			now:      time.Date(2026, 3, 24, 13, 0, 0, 0, time.UTC), // 2 PM Lagos
			expected: "2026-03-24 15:00:00",                    // 3 PM Lagos (Now + 1h)
		},
		{
			name:     "Late Night (11 PM Lagos)",
			now:      time.Date(2026, 3, 24, 22, 0, 0, 0, time.UTC), // 11 PM Lagos
			expected: "2026-03-25 08:00:00",                    // 8 AM Tomorrow
		},
		{
			name:     "Early Morning (5 AM Lagos)",
			now:      time.Date(2026, 3, 24, 4, 0, 0, 0, time.UTC), // 5 AM Lagos
			expected: "2026-03-24 08:00:00",                   // 8 AM Today
		},
		{
			name:     "Boundary 9:59 PM Lagos",
			now:      time.Date(2026, 3, 24, 20, 59, 0, 0, time.UTC), // 9:59 PM Lagos
			expected: "2026-03-25 08:00:00",                     // 10:59 PM (Hour 22) -> 8 AM Tomorrow
		},
	}

	loc, _ := time.LoadLocation(adminTZ)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := calc.CalculateDeparture(tt.now, adminTZ)
			assert.Equal(t, tt.expected, res.In(loc).Format("2006-01-02 15:04:05"))
		})
	}
}

func TestAlgorithmB_Arrival(t *testing.T) {
	calc := &Calculator{}
	// Departure: 12:00 PM UTC = 4:30 PM Kabul
	departure := time.Date(2026, 3, 24, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		origin   string
		dest     string
		expected string // Recipient Local Time
		minDur   int
		maxDur   int
	}{
		{
			origin: "Afghanistan",
			dest:   "Pakistan", // UTC+5
			minDur: 18,
			maxDur: 26,
		},
		{
			origin: "Nigeria",
			dest:   "USA", // UTC-5
			minDur: 30,
			maxDur: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.dest, func(t *testing.T) {
			arrival, _ := calc.CalculateArrival(departure, tt.origin, tt.dest)
			
			// Verify Local Window (9 AM - 4 PM)
			recipientTZ := calc.ResolveTimezone(tt.dest)
			loc, _ := time.LoadLocation(recipientTZ)
			localTime := arrival.In(loc)
			
			assert.True(t, localTime.Hour() >= 9 && localTime.Hour() <= 16, "Arrival at %v should be within 9-16", localTime)
		})
	}
}

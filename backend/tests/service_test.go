package tests

import (
	"testing"
	"time"
	"webtracker-bot/internal/shipment"

	"github.com/stretchr/testify/assert"
)

func TestAlgorithmA_Departure(t *testing.T) {
	calc := &shipment.Calculator{}

	tests := []struct {
		name          string
		originCountry string
		now           time.Time
		expected      string // Origin local time string
	}{
		{
			name:          "Normal Day Hour (2 PM Lagos)",
			originCountry: "Nigeria",
			now:           time.Date(2026, 3, 24, 13, 0, 0, 0, time.UTC), // 2 PM Lagos
			expected:      "2026-03-24 15:00:00",                         // 3 PM Lagos (Now + 1h)
		},
		{
			name:          "Late Night (11 PM Lagos)",
			originCountry: "Nigeria",
			now:           time.Date(2026, 3, 24, 22, 0, 0, 0, time.UTC), // 11 PM Lagos
			expected:      "2026-03-25 08:00:00",                         // 8 AM Next Day (capped)
		},
		{
			name:          "Early Morning (5 AM Lagos)",
			originCountry: "Nigeria",
			now:           time.Date(2026, 3, 24, 4, 0, 0, 0, time.UTC), // 5 AM Lagos
			expected:      "2026-03-24 08:00:00",                        // 8 AM Same Day (capped)
		},
		{
			name:          "Boundary 9:59 PM Lagos",
			originCountry: "Nigeria",
			now:           time.Date(2026, 3, 24, 20, 59, 0, 0, time.UTC), // 9:59 PM Lagos
			expected:      "2026-03-24 22:59:00",                          // 10:59 PM Lagos (within window, not capped)
		},
		{
			name:          "10 PM Lagos → 11 PM departure → capped",
			originCountry: "Nigeria",
			now:           time.Date(2026, 3, 24, 21, 0, 0, 0, time.UTC), // 10 PM Lagos
			expected:      "2026-03-25 08:00:00",                          // 11 PM (hour 23) → capped to 8 AM next day
		},
		{
			name:          "Afghanistan origin (3 PM Kabul)",
			originCountry: "Afghanistan",
			now:           time.Date(2026, 3, 24, 10, 30, 0, 0, time.UTC), // 3 PM Kabul (UTC+4:30)
			expected:      "2026-03-24 16:00:00",                          // 4:00 PM Kabul (Now + 1h)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := calc.CalculateDeparture(tt.now, tt.originCountry)
			tz := calc.ResolveTimezone(tt.originCountry)
			loc, _ := time.LoadLocation(tz)
			assert.Equal(t, tt.expected, res.In(loc).Format("2006-01-02 15:04:05"))
		})
	}
}

func TestAlgorithmB_Arrival(t *testing.T) {
	calc := &shipment.Calculator{}
	// Departure: 12:00 PM UTC
	departure := time.Date(2026, 3, 24, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		dest string
	}{
		{name: "Pakistan delivery", dest: "Pakistan"},
		{name: "USA delivery", dest: "USA"},
		{name: "Nigeria delivery", dest: "Nigeria"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrival, outForDelivery := calc.CalculateArrival(departure, tt.dest)

			// Verify destination local time
			recipientTZ := calc.ResolveTimezone(tt.dest)
			loc, _ := time.LoadLocation(recipientTZ)
			localArrival := arrival.In(loc)
			localOFD := outForDelivery.In(loc)

			// Expected Delivery: 9 AM - 5 PM local
			assert.True(t, localArrival.Hour() >= 9 && localArrival.Hour() <= 16,
				"Arrival at %v should be within 9-16", localArrival)

			// Out for Delivery: 7:30 AM - 9:00 AM local
			ofdMinutes := localOFD.Hour()*60 + localOFD.Minute()
			assert.True(t, ofdMinutes >= 7*60+30 && ofdMinutes <= 9*60,
				"Out for Delivery at %v should be within 7:30-9:00", localOFD)

			// Must be next day (not same day as departure)
			depLocal := departure.In(loc)
			assert.True(t, localArrival.Day() > depLocal.Day() || localArrival.Month() > depLocal.Month(),
				"Arrival must be after departure date")
		})
	}
}

func TestAlgorithmB_SundaySkip(t *testing.T) {
	calc := &shipment.Calculator{}

	// Saturday departure → arrival would be Sunday → should skip to Monday
	// 2026-03-28 is a Saturday
	saturdayDeparture := time.Date(2026, 3, 28, 10, 0, 0, 0, time.UTC)

	arrival, _ := calc.CalculateArrival(saturdayDeparture, "Nigeria")
	loc, _ := time.LoadLocation("Africa/Lagos")
	localArrival := arrival.In(loc)

	assert.Equal(t, time.Monday, localArrival.Weekday(),
		"Sunday arrival should be pushed to Monday, got %v", localArrival.Weekday())
}

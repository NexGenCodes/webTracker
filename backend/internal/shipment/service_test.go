package shipment

import (
	"testing"
	"time"
)

func TestResolveTimezone(t *testing.T) {
	calc := &Calculator{}

	tests := []struct {
		input    string
		expected string
	}{
		{"nigeria", "Africa/Lagos"},
		{"usa", "America/New_York"},
		{"unknown_country", "UTC"},
		{"china", "Asia/Shanghai"},
	}

	for _, tt := range tests {
		result := calc.ResolveTimezone(tt.input)
		if result != tt.expected {
			t.Errorf("ResolveTimezone(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestCalculateSchedule(t *testing.T) {
	calc := &Calculator{}

	// Fixed known time: 2023-10-25 12:00:00 UTC
	nowUTC := time.Date(2023, 10, 25, 12, 0, 0, 0, time.UTC)

	// Scenario 1: Same Timezone (UTC -> UTC)
	// 12:00 UTC -> Transit starts tomorrow 08:00 UTC (since 12 > 8 but < 21 ?? Wait, logic says:
	// if hour < 8: today 8am
	// else if hour >= 21: tomorrow 8am
	// else: now + 1hr
	// So 12:00 -> 13:00 start
	start, out, delivery := calc.CalculateSchedule(nowUTC, "unknown", "unknown")

	expectedStart := nowUTC.Add(1 * time.Hour)
	if !start.Equal(expectedStart) {
		t.Errorf("Scenario 1 Start: got %v, want %v", start, expectedStart)
	}

	// Delivery is Start + 24h, then reset to 10am
	// Start = 25th 13:00. +24h = 26th 13:00.
	// Reset to 10am = 26th 10:00.
	expectedDelivery := time.Date(2023, 10, 26, 10, 0, 0, 0, time.UTC)
	if !delivery.Equal(expectedDelivery) {
		t.Errorf("Scenario 1 Delivery: got %v, want %v", delivery, expectedDelivery)
	}

	// OutForDelivery is Delivery Day 8am
	expectedOut := time.Date(2023, 10, 26, 8, 0, 0, 0, time.UTC)
	if !out.Equal(expectedOut) {
		t.Errorf("Scenario 1 OutForDelivery: got %v, want %v", out, expectedOut)
	}
}

func TestCalculateSchedule_Night(t *testing.T) {
	calc := &Calculator{}
	// 2023-10-25 22:00:00 UTC (Night)
	nowUTC := time.Date(2023, 10, 25, 22, 0, 0, 0, time.UTC)

	// Logic: hour >= 21 -> Tomorrow 8am
	// Expected Start: 26th 08:00 UTC
	start, _, _ := calc.CalculateSchedule(nowUTC, "unknown", "unknown")

	expectedStart := time.Date(2023, 10, 26, 8, 0, 0, 0, time.UTC)
	if !start.Equal(expectedStart) {
		t.Errorf("Night Scenario Start: got %v, want %v", start, expectedStart)
	}
}

func TestCalculateSchedule_Morning(t *testing.T) {
	calc := &Calculator{}
	// 2023-10-25 05:00:00 UTC (Morning, before 8am)
	nowUTC := time.Date(2023, 10, 25, 5, 0, 0, 0, time.UTC)

	// Logic: hour < 8 -> Today 8am
	// Expected Start: 25th 08:00 UTC
	start, _, _ := calc.CalculateSchedule(nowUTC, "unknown", "unknown")

	expectedStart := time.Date(2023, 10, 25, 8, 0, 0, 0, time.UTC)
	if !start.Equal(expectedStart) {
		t.Errorf("Morning Scenario Start: got %v, want %v", start, expectedStart)
	}
}

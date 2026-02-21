package tests

import (
	"testing"
	"webtracker-bot/internal/parser"
)

func TestParseEditPairs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "Single field update",
			input: "name: John Doe",
			expected: map[string]string{
				"recipient_name": "John Doe",
			},
		},
		{
			name:  "Multi field update",
			input: "name: Mark, phone: 080123, departure: tomorrow",
			expected: map[string]string{
				"recipient_name":         "Mark",
				"recipient_phone":        "080123",
				"scheduled_transit_time": "tomorrow",
			},
		},
		{
			name:     "No labels (should return empty)",
			input:    "just some random text",
			expected: map[string]string{},
		},
		{
			name:  "Aliases",
			input: "to: Lagos, transit: yesterday",
			expected: map[string]string{
				"destination":            "Lagos",
				"scheduled_transit_time": "yesterday",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := parser.ParseEditPairs(tt.input)
			if len(results) != len(tt.expected) {
				t.Errorf("expected %d results, got %d", len(tt.expected), len(results))
			}
			for k, v := range tt.expected {
				if results[k] != v {
					t.Errorf("field %s: expected %s, got %s", k, v, results[k])
				}
			}
		})
	}
}

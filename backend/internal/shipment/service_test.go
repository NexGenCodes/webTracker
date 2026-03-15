package shipment

import (
	"testing"
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

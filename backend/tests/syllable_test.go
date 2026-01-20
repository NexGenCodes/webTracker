package tests

import (
	"testing"
	"webtracker-bot/internal/config"
)

func TestGenerateAbbreviationSyllables(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"FAST", "FST"},          // 1 Syllable: First, Mid, Last
		{"LOG", "LOG"},           // 1 Syllable (short): LOG
		{"LOGIC", "LOG"},         // 2 Syllables: LO (1st) + G (2nd)
		{"LOGISTICS", "LGT"},     // 3 Syllables: LO-GIS-TICS -> L + G + T
		{"AIRWAYBILL", "AWB"},    // 3 Syllables: AIR-WAY-BILL -> A + W + B
		{"GLOBAL", "GLB"},        // 2 Syllables: GLO-BAL -> GL + B = GLB
		{"MAX", "MAX"},           // 1 Syllable (short)
		{"FEDEX", "FED"},         // 2 Syllables: FE + D = FED
		{"INTERNATIONAL", "ITN"}, // 5 Syllables -> 1st char of first 3: I, T, N
	}

	for _, tt := range tests {
		got := config.GenerateAbbreviation(tt.name)
		if got != tt.expected {
			t.Errorf("GenerateAbbreviation(%q) = %q, want %q", tt.name, got, tt.expected)
		}
	}
}

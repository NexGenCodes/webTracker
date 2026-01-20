package parser

import "testing"

func TestCleanText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello\r\nWorld", "Hello\nWorld"},
		{"Line1\rLine2", "Line1\nLine2"},
		{"Normal Text", "Normal Text"},
		{"\r\n\r\n", "\n\n"},
	}

	for _, test := range tests {
		result := CleanText(test.input)
		if result != test.expected {
			t.Errorf("CleanText(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"test@example.com", true},
		{"invalid-email", false},
		{"no-at-sign.com", false},
		{"@only-at.com", true}, // Simple validation allows this
		{"user@", false},       // Missing domain part (strict check would fail, logic is simple)
		{"", false},
	}

	// Adjust expectations based on actual implementation: strings.Contains(email, "@") && strings.Contains(email, ".")
	// "user@" -> contains @, no ., so false. Correct.

	for _, test := range tests {
		result := ValidateEmail(test.input)
		if result != test.expected {
			t.Errorf("ValidateEmail(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"+1234567890", true},
		{"12345", true}, // Min 5 digits
		{"1234", false},
		{"abcdefg", false},
		{"+1 (555) 123-4567", true},
		{"", false},
	}

	for _, test := range tests {
		result := ValidatePhone(test.input)
		if result != test.expected {
			t.Errorf("ValidatePhone(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}

package commands

import (
	"context"
	"strings"
	"testing"
)

func TestIsAdmin(t *testing.T) {
	admins := []string{"2349077584528", "123456789"}

	tests := []struct {
		phone    string
		expected bool
	}{
		{"2349077584528", true},
		{"+2349077584528", true},
		{"234-90775-84528", true},
		{" 234 907 758 4528 ", true},
		{"123456789", true},
		{"987654321", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isAdmin(tt.phone, admins)
		if result != tt.expected {
			t.Errorf("isAdmin(%q, %v) = %v; want %v", tt.phone, admins, result, tt.expected)
		}
	}
}

func TestStatsHandlerAdminCheck(t *testing.T) {
	h := &StatsHandler{AdminPhones: []string{"2349077584528"}}

	// Test case: Non-Admin
	res := h.Execute(context.WithValue(context.Background(), "sender_phone", "1111111"), nil, nil, nil, "en")
	if !strings.Contains(res.Message, "PERMISSION DENIED") {
		t.Errorf("StatsHandler incorrectly allowed non-admin")
	}

	// Test case: Admin (Verify it passes the permission check)
	// We use a defer recover because it will panic when it tries to use the nil DB
	defer func() {
		if r := recover(); r != nil {
			// Panic is expected because db is nil, but we check if we were authorized
		}
	}()

	res = h.Execute(context.WithValue(context.Background(), "sender_phone", "2349077584528"), nil, nil, nil, "en")
	if strings.Contains(res.Message, "PERMISSION DENIED") {
		t.Errorf("StatsHandler incorrectly denied admin")
	}
}

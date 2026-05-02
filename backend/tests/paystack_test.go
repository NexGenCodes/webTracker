package tests

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"testing"
	"webtracker-bot/internal/payment"
)

func TestPaystackService_VerifySignature(t *testing.T) {
	secretKey := "sk_test_12345"
	service := payment.NewPaystackService(secretKey)

	payload := []byte(`{"event":"charge.success","data":{"id":123}}`)
	
	// Pre-calculate expected signature for test
	mac := hmac.New(sha512.New, []byte(secretKey))
	mac.Write(payload)
	validSignature := hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name      string
		payload   []byte
		signature string
		want      bool
	}{
		{
			name:      "Valid Signature",
			payload:   payload,
			signature: validSignature,
			want:      true,
		},
		{
			name:      "Invalid Signature",
			payload:   payload,
			signature: "invalid_sig",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := service.VerifySignature(tt.payload, tt.signature); got != tt.want {
				t.Errorf("PaystackService.VerifySignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPlanByID(t *testing.T) {
	tests := []struct {
		id      string
		wantID  string
		wantErr bool
	}{
		{"starter", "starter", false},
		{"pro", "pro", false},
		{"enterprise", "enterprise", false},
		{"scale", "enterprise", false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got, err := payment.GetPlanByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPlanByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("GetPlanByID() = %v, want %v", got.ID, tt.wantID)
			}
		})
	}
}

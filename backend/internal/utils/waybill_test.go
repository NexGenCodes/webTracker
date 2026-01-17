package utils

import (
	"strings"
	"testing"
	"time"
	"webtracker-bot/internal/models"
)

func TestGenerateWaybill(t *testing.T) {
	shipment := models.Shipment{
		TrackingNumber:  "AWB-TEST-123",
		Status:          "PENDING",
		SenderName:      "Sender Org",
		ReceiverName:    "Receiver Person",
		ReceiverPhone:   "08000000000",
		ReceiverAddress: "123 Waybill St",
		ReceiverCountry: "Nigeria",
		CreatedAt:       time.Now(),
	}

	got := GenerateWaybill(shipment, "TEST LOGISTICS")

	// Check for key elements
	if !strings.Contains(got, "AWB-TEST-123") {
		t.Errorf("Waybill missing tracking number")
	}
	if !strings.Contains(got, "Sender Org") {
		t.Errorf("Waybill missing sender name")
	}
	if !strings.Contains(got, "Receiver Person") {
		t.Errorf("Waybill missing receiver name")
	}
	if !strings.Contains(got, "=====") {
		t.Errorf("Waybill missing borders")
	}
}

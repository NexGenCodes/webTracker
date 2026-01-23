package utils

import (
	"fmt"
	"strings"
	"webtracker-bot/internal/shipment"
)

func GenerateWaybill(s shipment.Shipment, companyName string) string {
	var b strings.Builder

	if companyName == "" {
		companyName = "AIRWAY BILL LOGISTICS"
	}
	header := strings.ToUpper(companyName)

	border := "━━━━━━━━━━━━━━━━━━━━━━━"
	b.WriteString(border + "\n")
	padding := (len(border)/2 - len(header)/2) // Rough approximation for visual centering
	if padding < 0 {
		padding = 0
	}
	b.WriteString(strings.Repeat(" ", padding) + "⚡ " + header + "\n")
	b.WriteString(border + "\n\n")

	b.WriteString(fmt.Sprintf("🆔 [TRACKING ID]: %s\n", s.TrackingID))
	b.WriteString(fmt.Sprintf("📍 [STATUS]: %s\n", strings.ToUpper(s.Status)))
	b.WriteString(fmt.Sprintf("📅 [DATE]:   %s\n\n", s.CreatedAt.Format("02 Jan 2006, 15:04")))

	b.WriteString("👤 [SENDER INFORMATION]\n")
	b.WriteString(fmt.Sprintf("   • Name:    %s\n", s.SenderName))
	b.WriteString(fmt.Sprintf("   • Origin:  %s\n\n", s.Origin))

	b.WriteString("🎯 [RECEIVER INFORMATION]\n")
	b.WriteString(fmt.Sprintf("   • Name:    %s\n", s.RecipientName))
	b.WriteString(fmt.Sprintf("   • Phone:   %s\n", s.RecipientPhone))
	if s.RecipientEmail != "" {
		b.WriteString(fmt.Sprintf("   • Email:   %s\n", s.RecipientEmail))
	}
	if s.RecipientID != "" {
		b.WriteString(fmt.Sprintf("   • ID/PP:   %s\n", s.RecipientID))
	}
	b.WriteString(fmt.Sprintf("   • Destination: %s\n\n", s.Destination))

	b.WriteString(border + "\n")
	b.WriteString("  *THANK YOU FOR YOUR PATRONAGE*  \n")
	b.WriteString(border + "\n")

	return b.String()
}

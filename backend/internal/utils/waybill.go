package utils

import (
	"fmt"
	"strings"
	"webtracker-bot/internal/models"
)

func GenerateWaybill(s models.Shipment, companyName string) string {
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

	b.WriteString(fmt.Sprintf("🆔 [ID]:     %s\n", s.TrackingNumber))
	b.WriteString(fmt.Sprintf("📍 [STATUS]: %s\n", s.Status))
	b.WriteString(fmt.Sprintf("📅 [DATE]:   %s\n\n", s.CreatedAt.Format("02 Jan 2006, 15:04")))

	b.WriteString("👤 [SENDER INFORMATION]\n")
	b.WriteString(fmt.Sprintf("   • Name:    %s\n", s.SenderName))
	b.WriteString(fmt.Sprintf("   • Country: %s\n\n", s.SenderCountry))

	b.WriteString("🎯 [RECEIVER INFORMATION]\n")
	b.WriteString(fmt.Sprintf("   • Name:    %s\n", s.ReceiverName))
	b.WriteString(fmt.Sprintf("   • Phone:   %s\n", s.ReceiverPhone))
	if s.ReceiverEmail != "" {
		b.WriteString(fmt.Sprintf("   • Email:   %s\n", s.ReceiverEmail))
	}
	if s.ReceiverID != "" {
		b.WriteString(fmt.Sprintf("   • ID/NIN:  %s\n", s.ReceiverID))
	}
	b.WriteString(fmt.Sprintf("   • Address: %s\n", s.ReceiverAddress))
	b.WriteString(fmt.Sprintf("   • Country: %s\n\n", s.ReceiverCountry))

	b.WriteString(border + "\n")
	b.WriteString("  *THANK YOU FOR YOUR PATRONAGE*  \n")
	b.WriteString(border + "\n")

	return b.String()
}

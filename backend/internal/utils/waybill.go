package utils

import (
	"fmt"
	"strings"
	"webtracker-bot/internal/models"
)

func GenerateWaybill(s models.Shipment) string {
	var b strings.Builder

	border := "================================="
	b.WriteString(border + "\n")
	b.WriteString("       AIRWAY BILL LOGISTICS     \n")
	b.WriteString(border + "\n\n")

	b.WriteString(fmt.Sprintf("TRACKING: %s\n", s.TrackingNumber))
	b.WriteString(fmt.Sprintf("STATUS:   %s\n", s.Status))
	b.WriteString(fmt.Sprintf("DATE:     %s\n\n", s.CreatedAt.Format("2006-01-02 15:04")))

	b.WriteString("--- SENDER ---\n")
	b.WriteString(fmt.Sprintf("NAME:    %s\n", s.SenderName))
	b.WriteString(fmt.Sprintf("COUNTRY: %s\n\n", s.SenderCountry))

	b.WriteString("--- RECEIVER ---\n")
	b.WriteString(fmt.Sprintf("NAME:    %s\n", s.ReceiverName))
	b.WriteString(fmt.Sprintf("PHONE:   %s\n", s.ReceiverPhone))
	b.WriteString(fmt.Sprintf("ADDRESS: %s\n", s.ReceiverAddress))
	b.WriteString(fmt.Sprintf("COUNTRY: %s\n\n", s.ReceiverCountry))

	b.WriteString(border + "\n")
	b.WriteString("    THANK YOU FOR SHIPPING!    \n")
	b.WriteString(border + "\n")

	return b.String()
}

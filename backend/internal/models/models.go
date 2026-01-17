package models

import (
	"time"

	"go.mau.fi/whatsmeow/types"
)

type Manifest struct {
	ReceiverName    string   `json:"receiverName"`
	ReceiverAddress string   `json:"receiverAddress"`
	ReceiverPhone   string   `json:"receiverPhone"`
	ReceiverCountry string   `json:"receiverCountry"`
	SenderName      string   `json:"senderName"`
	SenderCountry   string   `json:"senderCountry"`
	MissingFields   []string `json:"-"`
}

type Shipment struct {
	ID              string    `json:"id"`
	TrackingNumber  string    `json:"trackingNumber"`
	Status          string    `json:"status"`
	SenderName      string    `json:"senderName"`
	SenderCountry   string    `json:"senderCountry"`
	ReceiverName    string    `json:"receiverName"`
	ReceiverPhone   string    `json:"receiverPhone"`
	ReceiverAddress string    `json:"receiverAddress"`
	ReceiverCountry string    `json:"receiverCountry"`
	CreatedAt       time.Time `json:"createdAt"`
}

type Job struct {
	ChatJID     types.JID
	MessageID   string
	Text        string
	SenderPhone string
}

func StrPtr(s string) *string {
	return &s
}

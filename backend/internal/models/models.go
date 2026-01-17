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
	ReceiverEmail   string   `json:"receiverEmail"`
	ReceiverID      string   `json:"receiverID"`
	SenderName      string   `json:"senderName"`
	SenderCountry   string   `json:"senderCountry"`
	MissingFields   []string `json:"-"`
}

type Shipment struct {
	ID              string     `json:"id"`
	TrackingNumber  string     `json:"trackingNumber"`
	Status          string     `json:"status"`
	SenderName      string     `json:"senderName"`
	SenderCountry   string     `json:"senderCountry"`
	ReceiverName    string     `json:"receiverName"`
	ReceiverPhone   string     `json:"receiverPhone"`
	ReceiverEmail   string     `json:"receiverEmail"`
	ReceiverID      string     `json:"receiverID"`
	ReceiverAddress string     `json:"receiverAddress"`
	ReceiverCountry string     `json:"receiverCountry"`
	WhatsappFrom    string     `json:"whatsappFrom"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	LastNotifiedAt  *time.Time `json:"lastNotifiedAt"`
}

type NotificationJob struct {
	TrackingNumber string
	Status         string
	WhatsappFrom   string
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

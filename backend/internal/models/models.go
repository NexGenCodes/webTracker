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
	IsAI            bool     `json:"-"`
	MissingFields   []string `json:"-"`
}

// Merge combines this manifest with another, only filling in empty fields.
func (m *Manifest) Merge(other Manifest) {
	if m.ReceiverName == "" {
		m.ReceiverName = other.ReceiverName
	}
	if m.ReceiverAddress == "" {
		m.ReceiverAddress = other.ReceiverAddress
	}
	if m.ReceiverPhone == "" {
		m.ReceiverPhone = other.ReceiverPhone
	}
	if m.ReceiverCountry == "" {
		m.ReceiverCountry = other.ReceiverCountry
	}
	if m.ReceiverEmail == "" {
		m.ReceiverEmail = other.ReceiverEmail
	}
	if m.ReceiverID == "" {
		m.ReceiverID = other.ReceiverID
	}
	if m.SenderName == "" {
		m.SenderName = other.SenderName
	}
	if m.SenderCountry == "" {
		m.SenderCountry = other.SenderCountry
	}
}

// Validate checks for required fields and returns a list of missing ones.
func (m *Manifest) Validate() []string {
	var missing []string
	if m.ReceiverName == "" {
		missing = append(missing, "Receiver Name")
	}
	if m.ReceiverPhone == "" {
		missing = append(missing, "Receiver Phone")
	}
	if m.ReceiverAddress == "" {
		missing = append(missing, "Receiver Address")
	}
	if m.ReceiverCountry == "" {
		missing = append(missing, "Receiver Country")
	}
	if m.SenderName == "" {
		missing = append(missing, "Sender Name")
	}
	if m.SenderCountry == "" {
		missing = append(missing, "Sender Country")
	}
	m.MissingFields = missing
	return missing
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
	WhatsappFrom    *string    `json:"whatsappFrom"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	LastNotifiedAt  *time.Time `json:"lastNotifiedAt"`
}

type NotificationJob struct {
	TrackingNumber string
	Status         string
	WhatsappFrom   *string
}

type Job struct {
	ChatJID     types.JID
	SenderJID   types.JID
	MessageID   string
	Text        string
	SenderPhone string
}

func StrPtr(s string) *string {
	return &s
}

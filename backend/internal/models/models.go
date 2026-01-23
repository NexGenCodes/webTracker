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
	fillIfEmpty(&m.ReceiverName, other.ReceiverName)
	fillIfEmpty(&m.ReceiverAddress, other.ReceiverAddress)
	fillIfEmpty(&m.ReceiverPhone, other.ReceiverPhone)
	fillIfEmpty(&m.ReceiverCountry, other.ReceiverCountry)
	fillIfEmpty(&m.ReceiverEmail, other.ReceiverEmail)
	fillIfEmpty(&m.ReceiverID, other.ReceiverID)
	fillIfEmpty(&m.SenderName, other.SenderName)
	fillIfEmpty(&m.SenderCountry, other.SenderCountry)
}

func fillIfEmpty(target *string, val string) {
	if *target == "" {
		*target = val
	}
}

// Validate checks for required fields and returns a list of missing ones.
func (m *Manifest) Validate() []string {
	var missing []string
	check := func(val, name string) {
		if val == "" {
			missing = append(missing, name)
		}
	}

	check(m.ReceiverName, "Receiver Name")
	check(m.ReceiverPhone, "Receiver Phone")
	check(m.ReceiverAddress, "Receiver Address")
	check(m.ReceiverCountry, "Receiver Country")
	check(m.SenderName, "Sender Name")
	check(m.SenderCountry, "Sender Country")

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

type UserPreference struct {
	JID      string `json:"jid"`
	Language string `json:"language"`
}

type Job struct {
	ChatJID     types.JID
	SenderJID   types.JID
	MessageID   string
	Text        string
	SenderPhone string
	Language    string
	IsAdmin     bool
}

func StrPtr(s string) *string {
	return &s
}

func Uint64Ptr(u uint64) *uint64 {
	return &u
}

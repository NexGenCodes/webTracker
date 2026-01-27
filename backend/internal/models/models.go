package models

import (
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
	CargoType       string   `json:"cargoType"`
	Weight          float64  `json:"weight"`
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
	if m.Weight == 0 && other.Weight > 0 {
		m.Weight = other.Weight
	}
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

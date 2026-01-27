package shipment

import (
	"time"
)

// Status constants
const (
	StatusPending        = "pending"
	StatusIntransit      = "intransit"
	StatusOutForDelivery = "outfordelivery"
	StatusDelivered      = "delivered"
	StatusCanceled       = "canceled"
)

// Shipment represents the core data model for a package
type Shipment struct {
	// Unique Identifier
	TrackingID string `json:"tracking_id"`
	UserJID    string `json:"-"` // Hidden from public API

	// Current State
	Status string `json:"status"`

	// Core Timestamps (UTC)
	CreatedAt            time.Time `json:"created_at"`
	ScheduledTransitTime time.Time `json:"scheduled_transit_time"` // When it goes intransit
	OutForDeliveryTime   time.Time `json:"out_for_delivery_time"`  // Added for Arrival Notification
	ExpectedDeliveryTime time.Time `json:"expected_delivery_time"` // When it gets delivered

	// Timezone Metadata (Used for display logic)
	SenderTimezone    string `json:"sender_timezone"`
	RecipientTimezone string `json:"recipient_timezone"`

	// Shipment Details
	SenderName       string  `json:"sender_name"`
	SenderPhone      string  `json:"sender_phone"`
	Origin           string  `json:"origin"`
	RecipientName    string  `json:"recipient_name"`
	RecipientPhone   string  `json:"recipient_phone"`
	RecipientID      string  `json:"recipient_id"`
	RecipientEmail   string  `json:"recipient_email"`
	RecipientAddress string  `json:"recipient_address"`
	Destination      string  `json:"destination"`
	CargoType        string  `json:"cargo_type"`
	Weight           float64 `json:"weight"`
	Cost             float64 `json:"cost"`
}

// TimelineEvent represents a single step in the tracking history
type TimelineEvent struct {
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
	IsCompleted bool      `json:"is_completed"`
}

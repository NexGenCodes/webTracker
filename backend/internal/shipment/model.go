package shipment

import (
	"time"
	"webtracker-bot/internal/database/db"
)

// ToDomain converts a database Shipment model into the domain Shipment struct.
func ToDomain(dbShip db.Shipment) Shipment {
	var scheduledTransit, outForDelivery, expectedDelivery *time.Time
	if dbShip.ScheduledTransitTime.Valid {
		scheduledTransit = &dbShip.ScheduledTransitTime.Time
	}
	if dbShip.OutfordeliveryTime.Valid {
		outForDelivery = &dbShip.OutfordeliveryTime.Time
	}
	if dbShip.ExpectedDeliveryTime.Valid {
		expectedDelivery = &dbShip.ExpectedDeliveryTime.Time
	}

	return Shipment{
		TrackingID:           dbShip.TrackingID,
		UserJID:              dbShip.UserJid,
		Status:               dbShip.Status.String,
		CreatedAt:            dbShip.CreatedAt.Time,
		ScheduledTransitTime: scheduledTransit,
		OutForDeliveryTime:   outForDelivery,
		ExpectedDeliveryTime: expectedDelivery,
		SenderTimezone:       dbShip.SenderTimezone.String,
		RecipientTimezone:    dbShip.RecipientTimezone.String,
		SenderName:           dbShip.SenderName.String,
		SenderPhone:          dbShip.SenderPhone.String,
		Origin:               dbShip.Origin.String,
		RecipientName:        dbShip.RecipientName.String,
		RecipientPhone:       dbShip.RecipientPhone.String,
		RecipientID:          dbShip.RecipientID.String,
		RecipientEmail:       dbShip.RecipientEmail.String,
		RecipientAddress:     dbShip.RecipientAddress.String,
		Destination:          dbShip.Destination.String,
		CargoType:            dbShip.CargoType.String,
		Weight:               dbShip.Weight.Float64,
		Cost:                 dbShip.Cost.Float64,
	}
}

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
	CreatedAt            time.Time  `json:"created_at"`
	ScheduledTransitTime *time.Time `json:"scheduled_transit_time"` // When it goes intransit
	OutForDeliveryTime   *time.Time `json:"out_for_delivery_time"`  // Added for Arrival Notification
	ExpectedDeliveryTime *time.Time `json:"expected_delivery_time"` // When it gets delivered

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

// ResolveStatus returns what the status *should* be right now based on the schedule.
func (s *Shipment) ResolveStatus(nowUTC time.Time) string {
	if s.Status == StatusCanceled {
		return StatusCanceled
	}
	if s.ExpectedDeliveryTime != nil && !nowUTC.Before(*s.ExpectedDeliveryTime) {
		return StatusDelivered
	}
	if s.OutForDeliveryTime != nil && !nowUTC.Before(*s.OutForDeliveryTime) {
		return StatusOutForDelivery
	}
	if s.ScheduledTransitTime != nil && !nowUTC.Before(*s.ScheduledTransitTime) {
		return StatusIntransit
	}
	return StatusPending
}

// TimelineEvent represents a single step in the tracking history
type TimelineEvent struct {
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
	IsCompleted bool      `json:"is_completed"`
}

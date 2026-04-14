package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"webtracker-bot/internal/adapter/db"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/utils/dbutil"
	"github.com/sqlc-dev/pqtype"
)

func toNullString(s string) sql.NullString {
	return dbutil.ToNullString(s)
}

func toNullTime(t time.Time) sql.NullTime {
	return dbutil.ToNullTime(t)
}

// ShipmentUsecase exposes business logic operations for shipments.
type ShipmentUsecase struct {
	repo    db.Querier
	Service shipment.Service // Added
}

// NewShipmentUsecase creates a new usecase layer with the given repository and service.
func NewShipmentUsecase(repo db.Querier, service shipment.Service) *ShipmentUsecase {
	return &ShipmentUsecase{
		repo:    repo,
		Service: service,
	}
}

// Track retrieves a shipment by its tracking ID.
func (u *ShipmentUsecase) Track(ctx context.Context, trackingID string) (*db.Shipment, error) {
	shipment, err := u.repo.GetShipment(ctx, trackingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipment: %w", err)
	}
	return &shipment, nil
}

// Create inserts a new shipment into the system.
func (u *ShipmentUsecase) Create(ctx context.Context, params db.CreateShipmentParams) error {
	// Add business rules here (e.g., validation)
	if params.TrackingID == "" {
		return fmt.Errorf("tracking ID is required")
	}
	
	err := u.repo.CreateShipment(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create shipment: %w", err)
	}
	return nil
}

// UpdateStatus changes the status and destination of a shipment.
func (u *ShipmentUsecase) UpdateStatus(ctx context.Context, trackingID, status, destination string) error {
	params := db.UpdateShipmentStatusParams{
		TrackingID:  trackingID,
		Status:      toNullString(status),
		Destination: toNullString(destination),
	}
	err := u.repo.UpdateShipmentStatus(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

// ListPaginated returns a list of shipments with pagination.
func (u *ShipmentUsecase) ListPaginated(ctx context.Context, limit, offset int32) ([]db.Shipment, error) {
	params := db.ListShipmentsParams{
		Limit:  limit,
		Offset: offset,
	}
	shipments, err := u.repo.ListShipments(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list shipments: %w", err)
	}
	return shipments, nil
}

// Delete permanently removes a shipment record.
func (u *ShipmentUsecase) Delete(ctx context.Context, trackingID string) error {
	err := u.repo.DeleteShipment(ctx, trackingID)
	if err != nil {
		return fmt.Errorf("failed to delete shipment: %w", err)
	}
	return nil
}

// DeleteDelivered performs bulk cleanup of delivered shipments.
func (u *ShipmentUsecase) DeleteDelivered(ctx context.Context) error {
	err := u.repo.DeleteDeliveredShipments(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete delivered shipments: %w", err)
	}
	return nil
}

// ProcessMaintenance transitions shipments strictly based on their scheduled times.
// Returns the count of transitioned shipments.
func (u *ShipmentUsecase) ProcessMaintenance(ctx context.Context, now time.Time) (int, error) {
	totalTransitioned := 0

	// 1. Pending -> Intransit
	transit, err := u.repo.TransitionStatusToIntransit(ctx, toNullTime(now))
	if err != nil {
		return 0, fmt.Errorf("failed to transition to intransit: %w", err)
	}
	totalTransitioned += len(transit)

	// 2. Intransit -> OutForDelivery
	out, err := u.repo.TransitionStatusToOutForDelivery(ctx, toNullTime(now))
	if err != nil {
		return totalTransitioned, fmt.Errorf("failed to transition to outfordelivery: %w", err)
	}
	totalTransitioned += len(out)

	// 3. OutForDelivery -> Delivered
	delivered, err := u.repo.TransitionStatusToDelivered(ctx, toNullTime(now))
	if err != nil {
		return totalTransitioned, fmt.Errorf("failed to transition to delivered: %w", err)
	}
	totalTransitioned += len(delivered)

	return totalTransitioned, nil
}

// RunAgedCleanup deletes shipments that were delivered very long ago or created very long ago.
func (u *ShipmentUsecase) RunAgedCleanup(ctx context.Context, deliveredOlderThan, createdOlderThan time.Time) (int64, error) {
	params := db.RunAgedCleanupParams{
		UpdatedAt: toNullTime(deliveredOlderThan),
		CreatedAt: toNullTime(createdOlderThan),
	}
	err := u.repo.RunAgedCleanup(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("failed to run aged cleanup: %w", err)
	}
	return 0, nil // Count of deleted rows (not returned by :exec)
}

// CreateWithPrefix generates a tracking ID and inserts a new shipment.
func (u *ShipmentUsecase) CreateWithPrefix(ctx context.Context, s *db.Shipment, prefix string) (string, error) {
	if prefix == "" {
		prefix = "AWB"
	}
	seed := time.Now().UnixNano()
	randStr := fmt.Sprintf("%09d", seed%1000000000)
	trackingID := fmt.Sprintf("%s-%s", prefix, randStr)

	params := db.CreateShipmentParams{
		TrackingID:           trackingID,
		UserJid:              s.UserJid,
		Status:               s.Status,
		CreatedAt:            sql.NullTime{Time: time.Now(), Valid: true},
		ScheduledTransitTime: s.ScheduledTransitTime,
		OutfordeliveryTime:   s.OutfordeliveryTime,
		ExpectedDeliveryTime: s.ExpectedDeliveryTime,
		SenderTimezone:       s.SenderTimezone,
		RecipientTimezone:    s.RecipientTimezone,
		SenderName:           s.SenderName,
		SenderPhone:          s.SenderPhone,
		Origin:               s.Origin,
		RecipientName:        s.RecipientName,
		RecipientPhone:       s.RecipientPhone,
		RecipientEmail:       s.RecipientEmail,
		RecipientID:          s.RecipientID,
		RecipientAddress:     s.RecipientAddress,
		Destination:          s.Destination,
		CargoType:            s.CargoType,
		Weight:               s.Weight,
		Cost:                 s.Cost,
		UpdatedAt:            sql.NullTime{Time: time.Now(), Valid: true},
	}

	err := u.repo.CreateShipment(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create shipment: %w", err)
	}
	return trackingID, nil
}

// FindSimilar checks if a shipment already exists with matching recipient details.
func (u *ShipmentUsecase) FindSimilar(ctx context.Context, userJID, phone string) (string, error) {
	id, err := u.repo.FindSimilarShipment(ctx, db.FindSimilarShipmentParams{
		UserJid:        userJID,
		RecipientPhone: toNullString(phone),
	})
	if err == sql.ErrNoRows {
		return "", nil
	}
	return id, err
}

// GetLastForUser returns the tracking ID of the most recently created shipment for a user.
func (u *ShipmentUsecase) GetLastForUser(ctx context.Context, userJID string) (string, error) {
	id, err := u.repo.GetLastShipmentIDForUser(ctx, userJID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return id, err
}

// UpdateField updates a single field on a shipment by name.
func (u *ShipmentUsecase) UpdateField(ctx context.Context, trackingID, field, value string) error {
	switch field {
	case "sender_name":
		return u.repo.UpdateShipmentFieldSenderName(ctx, db.UpdateShipmentFieldSenderNameParams{TrackingID: trackingID, SenderName: toNullString(value)})
	case "sender_phone":
		return u.repo.UpdateShipmentFieldSenderPhone(ctx, db.UpdateShipmentFieldSenderPhoneParams{TrackingID: trackingID, SenderPhone: toNullString(value)})
	case "origin":
		return u.repo.UpdateShipmentFieldOrigin(ctx, db.UpdateShipmentFieldOriginParams{TrackingID: trackingID, Origin: toNullString(value)})
	case "recipient_name":
		return u.repo.UpdateShipmentFieldRecipientName(ctx, db.UpdateShipmentFieldRecipientNameParams{TrackingID: trackingID, RecipientName: toNullString(value)})
	case "recipient_phone":
		return u.repo.UpdateShipmentFieldRecipientPhone(ctx, db.UpdateShipmentFieldRecipientPhoneParams{TrackingID: trackingID, RecipientPhone: toNullString(value)})
	case "recipient_email":
		return u.repo.UpdateShipmentFieldRecipientEmail(ctx, db.UpdateShipmentFieldRecipientEmailParams{TrackingID: trackingID, RecipientEmail: toNullString(value)})
	case "recipient_id":
		return u.repo.UpdateShipmentFieldRecipientID(ctx, db.UpdateShipmentFieldRecipientIDParams{TrackingID: trackingID, RecipientID: toNullString(value)})
	case "recipient_address":
		return u.repo.UpdateShipmentFieldRecipientAddress(ctx, db.UpdateShipmentFieldRecipientAddressParams{TrackingID: trackingID, RecipientAddress: toNullString(value)})
	case "destination":
		return u.repo.UpdateShipmentFieldDestination(ctx, db.UpdateShipmentFieldDestinationParams{TrackingID: trackingID, Destination: toNullString(value)})
	case "cargo_type":
		return u.repo.UpdateShipmentFieldCargoType(ctx, db.UpdateShipmentFieldCargoTypeParams{TrackingID: trackingID, CargoType: toNullString(value)})
	case "scheduled_transit_time":
		t, err := time.Parse("2006-01-02 15:04:05", value)
		if err != nil {
			return fmt.Errorf("invalid time format: %w", err)
		}
		return u.repo.UpdateShipmentFieldScheduledTransitTime(ctx, db.UpdateShipmentFieldScheduledTransitTimeParams{TrackingID: trackingID, ScheduledTransitTime: toNullTime(t)})
	case "expected_delivery_time":
		t, err := time.Parse("2006-01-02 15:04:05", value)
		if err != nil {
			return fmt.Errorf("invalid time format: %w", err)
		}
		return u.repo.UpdateShipmentFieldExpectedDeliveryTime(ctx, db.UpdateShipmentFieldExpectedDeliveryTimeParams{TrackingID: trackingID, ExpectedDeliveryTime: toNullTime(t)})
	case "outfordelivery_time":
		t, err := time.Parse("2006-01-02 15:04:05", value)
		if err != nil {
			return fmt.Errorf("invalid time format: %w", err)
		}
		return u.repo.UpdateShipmentFieldOutfordeliveryTime(ctx, db.UpdateShipmentFieldOutfordeliveryTimeParams{TrackingID: trackingID, OutfordeliveryTime: toNullTime(t)})
	default:
		return fmt.Errorf("unsupported field: %s", field)
	}
}

// CountDailyStats returns the count of created and delivered shipments since the given time.
func (u *ShipmentUsecase) CountDailyStats(ctx context.Context, since time.Time) (created int64, delivered int64, err error) {
	created, err = u.repo.CountCreatedSince(ctx, toNullTime(since))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count created: %w", err)
	}
	delivered, err = u.repo.CountDeliveredSince(ctx, toNullTime(since))
	if err != nil {
		return created, 0, fmt.Errorf("failed to count delivered: %w", err)
	}
	return created, delivered, nil
}

// ListAll returns all shipments ordered by creation date.
func (u *ShipmentUsecase) ListAll(ctx context.Context) ([]db.Shipment, error) {
	return u.repo.ListAllShipments(ctx)
}

// TransitionResult represents a single status transition for notification purposes.
type TransitionResult struct {
	TrackingID     string
	NewStatus      string
	UserJID        string
	RecipientEmail string
}

// ProcessTransitions runs all status transitions and returns the results for notification.
func (u *ShipmentUsecase) ProcessTransitions(ctx context.Context, now time.Time) ([]TransitionResult, error) {
	var results []TransitionResult

	transit, err := u.repo.TransitionStatusToIntransit(ctx, toNullTime(now))
	if err != nil {
		return nil, err
	}
	for _, t := range transit {
		results = append(results, TransitionResult{TrackingID: t.TrackingID, NewStatus: t.NewStatus.String, UserJID: t.UserJid, RecipientEmail: t.RecipientEmail.String})
	}

	out, err := u.repo.TransitionStatusToOutForDelivery(ctx, toNullTime(now))
	if err != nil {
		return results, err
	}
	for _, t := range out {
		results = append(results, TransitionResult{TrackingID: t.TrackingID, NewStatus: t.NewStatus.String, UserJID: t.UserJid, RecipientEmail: t.RecipientEmail.String})
	}

	delivered, err := u.repo.TransitionStatusToDelivered(ctx, toNullTime(now))
	if err != nil {
		return results, err
	}
	for _, t := range delivered {
		results = append(results, TransitionResult{TrackingID: t.TrackingID, NewStatus: t.NewStatus.String, UserJID: t.UserJid, RecipientEmail: t.RecipientEmail.String})
	}

	return results, nil
}

// CountAll returns the total number of shipments.
func (u *ShipmentUsecase) CountAll(ctx context.Context) (int64, error) {
	return u.repo.CountShipments(ctx)
}

// CountByStatus returns shipment counts broken down by status.
func (u *ShipmentUsecase) CountByStatus(ctx context.Context) (*db.CountShipmentsByStatusRow, error) {
	stats, err := u.repo.CountShipmentsByStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count by status: %w", err)
	}
	return &stats, nil
}

// RecordEvent logs a system event for analytics.
func (u *ShipmentUsecase) RecordEvent(ctx context.Context, eventType string, metadata []byte) error {
	return u.repo.RecordEvent(ctx, db.RecordEventParams{
		EventType: eventType,
		Metadata:  pqtype.NullRawMessage{RawMessage: json.RawMessage(metadata), Valid: len(metadata) > 0},
	})
}

// GetTelemetryStats retrieves event counts categorized by type since a specific time.
func (u *ShipmentUsecase) GetTelemetryStats(ctx context.Context, since time.Time) ([]db.GetTelemetryStatsRow, error) {
	return u.repo.GetTelemetryStats(ctx, toNullTime(since))
}

// GetRecentEvents retrieves the latest telemetry logs.
func (u *ShipmentUsecase) GetRecentEvents(ctx context.Context, limit int32) ([]db.Telemetry, error) {
	return u.repo.GetRecentEvents(ctx, limit)
}

// BulkUpdateStatus updates the status of multiple shipments at once.
func (u *ShipmentUsecase) BulkUpdateStatus(ctx context.Context, ids []string, status string) error {
	return u.repo.BulkUpdateStatus(ctx, db.BulkUpdateStatusParams{
		Column1: ids,
		Status:  toNullString(status),
	})
}

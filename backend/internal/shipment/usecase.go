package shipment

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/database/dbutil"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/utils"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

func toNullUUID(id uuid.UUID) uuid.NullUUID {
	return uuid.NullUUID{UUID: id, Valid: true}
}

// ShipmentUsecase exposes business logic operations for shipments.
type Usecase struct {
	repo    db.Querier
	Service Service
}

// NewUsecase creates a new usecase layer with the given repository and service.
func NewUsecase(repo db.Querier, service Service) *Usecase {
	return &Usecase{
		repo:    repo,
		Service: service,
	}
}

func (u *Usecase) GetService() models.ShipmentService {
	return u.Service
}

// Track retrieves a shipment by its tracking ID.
func (u *Usecase) Track(ctx context.Context, companyID uuid.UUID, trackingID string) (*db.Shipment, error) {
	shipment, err := u.repo.GetShipment(ctx, db.GetShipmentParams{CompanyID: toNullUUID(companyID), TrackingID: trackingID})
	if err != nil {
		return nil, fmt.Errorf("failed to get shipment: %w", err)
	}
	return &shipment, nil
}

// Create inserts a new shipment into the system.
func (u *Usecase) Create(ctx context.Context, companyID uuid.UUID, params db.CreateShipmentParams) error {
	if params.TrackingID == "" {
		return fmt.Errorf("tracking ID is required")
	}
	params.CompanyID = toNullUUID(companyID)
	err := u.repo.CreateShipment(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create shipment: %w", err)
	}
	return nil
}

// UpdateStatus changes the status and destination of a shipment.
func (u *Usecase) UpdateStatus(ctx context.Context, companyID uuid.UUID, trackingID, status, destination string) error {
	params := db.UpdateShipmentStatusParams{
		CompanyID:   toNullUUID(companyID),
		TrackingID:  trackingID,
		Status:      dbutil.ToNullString(status),
		Destination: dbutil.ToNullString(destination),
	}
	err := u.repo.UpdateShipmentStatus(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

// ListPaginated returns a list of shipments with pagination.
func (u *Usecase) ListPaginated(ctx context.Context, companyID uuid.UUID, limit, offset int32) ([]db.Shipment, error) {
	params := db.ListShipmentsParams{
		CompanyID: toNullUUID(companyID),
		Limit:     limit,
		Offset:    offset,
	}
	shipments, err := u.repo.ListShipments(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list shipments: %w", err)
	}
	return shipments, nil
}

// Delete permanently removes a shipment record.
func (u *Usecase) Delete(ctx context.Context, companyID uuid.UUID, trackingID string) error {
	err := u.repo.DeleteShipment(ctx, db.DeleteShipmentParams{CompanyID: toNullUUID(companyID), TrackingID: trackingID})
	if err != nil {
		return fmt.Errorf("failed to delete shipment: %w", err)
	}
	return nil
}

// BulkDelete removes multiple shipments in a single DB round-trip.
func (u *Usecase) BulkDelete(ctx context.Context, companyID uuid.UUID, ids []string) (int64, error) {
	result, err := u.repo.BulkDeleteShipments(ctx, db.BulkDeleteShipmentsParams{
		CompanyID: toNullUUID(companyID),
		Column2:   ids,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to bulk delete shipments: %w", err)
	}
	deleted, _ := result.RowsAffected()
	return deleted, nil
}

// DeleteDelivered performs bulk cleanup of delivered shipments.
func (u *Usecase) DeleteDelivered(ctx context.Context, companyID uuid.UUID) error {
	err := u.repo.DeleteDeliveredShipments(ctx, toNullUUID(companyID))
	if err != nil {
		return fmt.Errorf("failed to delete delivered shipments: %w", err)
	}
	return nil
}

// ProcessMaintenance transitions shipments strictly based on their scheduled times.
func (u *Usecase) ProcessMaintenance(ctx context.Context, companyID uuid.UUID, now time.Time) (int, error) {
	totalTransitioned := 0

	transit, err := u.repo.TransitionStatusToIntransit(ctx, db.TransitionStatusToIntransitParams{CompanyID: toNullUUID(companyID), ScheduledTransitTime: dbutil.ToNullTime(now)})
	if err != nil {
		return 0, fmt.Errorf("failed to transition to intransit: %w", err)
	}
	totalTransitioned += len(transit)

	out, err := u.repo.TransitionStatusToOutForDelivery(ctx, db.TransitionStatusToOutForDeliveryParams{CompanyID: toNullUUID(companyID), OutfordeliveryTime: dbutil.ToNullTime(now)})
	if err != nil {
		return totalTransitioned, fmt.Errorf("failed to transition to outfordelivery: %w", err)
	}
	totalTransitioned += len(out)

	delivered, err := u.repo.TransitionStatusToDelivered(ctx, db.TransitionStatusToDeliveredParams{CompanyID: toNullUUID(companyID), ExpectedDeliveryTime: dbutil.ToNullTime(now)})
	if err != nil {
		return totalTransitioned, fmt.Errorf("failed to transition to delivered: %w", err)
	}
	totalTransitioned += len(delivered)

	return totalTransitioned, nil
}

// RunAgedCleanup deletes shipments that were delivered very long ago or created very long ago.
func (u *Usecase) RunAgedCleanup(ctx context.Context, companyID uuid.UUID, deliveredOlderThan, createdOlderThan time.Time) (int64, error) {
	params := db.RunAgedCleanupParams{
		CompanyID: toNullUUID(companyID),
		UpdatedAt: dbutil.ToNullTime(deliveredOlderThan),
		CreatedAt: dbutil.ToNullTime(createdOlderThan),
	}
	result, err := u.repo.RunAgedCleanup(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("failed to run aged cleanup: %w", err)
	}
	deleted, _ := result.RowsAffected()
	return deleted, nil
}

// CreateWithPrefix generates a tracking ID and inserts a new shipment.
func (u *Usecase) CreateWithPrefix(ctx context.Context, companyID uuid.UUID, s *db.Shipment, prefix string) (string, error) {
	if prefix == "" {
		prefix = "AWB"
	}

	var trackingID string
	var err error

	for attempts := 0; attempts < 5; attempts++ {
		trackingID, err = utils.GenerateTrackingID(prefix)
		if err != nil {
			return "", err
		}

		params := db.CreateShipmentParams{
			CompanyID:            toNullUUID(companyID),
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

		err = u.repo.CreateShipment(ctx, params)
		if err == nil {
			return trackingID, nil
		}

		if !strings.Contains(err.Error(), "duplicate key value") && !strings.Contains(err.Error(), "23505") {
			return "", fmt.Errorf("failed to create shipment: %w", err)
		}
	}

	return "", fmt.Errorf("failed to create shipment after 5 retries due to ID collision: %w", err)
}

// FindSimilar checks if a shipment already exists with matching recipient details.
func (u *Usecase) FindSimilar(ctx context.Context, companyID uuid.UUID, userJID, phone string) (string, error) {
	id, err := u.repo.FindSimilarShipment(ctx, db.FindSimilarShipmentParams{
		CompanyID:      toNullUUID(companyID),
		UserJid:        userJID,
		RecipientPhone: dbutil.ToNullString(phone),
	})
	if err == sql.ErrNoRows {
		return "", nil
	}
	return id, err
}

// GetLastForUser returns the tracking ID of the most recently created shipment for a user.
func (u *Usecase) GetLastForUser(ctx context.Context, companyID uuid.UUID, userJID string) (string, error) {
	id, err := u.repo.GetLastShipmentIDForUser(ctx, db.GetLastShipmentIDForUserParams{CompanyID: toNullUUID(companyID), UserJid: userJID})
	if err == sql.ErrNoRows {
		return "", nil
	}
	return id, err
}

// UpdateField updates a single field on a shipment by name.
func (u *Usecase) UpdateField(ctx context.Context, companyID uuid.UUID, trackingID, field, value string) error {
	params := db.UpdateShipmentDynamicParams{
		CompanyID:  toNullUUID(companyID),
		TrackingID: trackingID,
	}

	switch field {
	case "sender_name":
		params.Column3 = value
	case "sender_phone":
		params.Column4 = value
	case "origin":
		params.Column5 = value
	case "recipient_name":
		params.Column6 = value
	case "recipient_phone":
		params.Column7 = value
	case "recipient_email":
		params.Column8 = value
	case "recipient_id":
		params.Column9 = value
	case "recipient_address":
		params.Column10 = value
	case "destination":
		params.Column11 = value
	case "cargo_type":
		params.Column12 = value
	case "scheduled_transit_time":
		t, err := parseFlexibleTime(value)
		if err != nil {
			return fmt.Errorf("invalid time format: %w", err)
		}
		params.Column13 = t
	case "expected_delivery_time":
		t, err := parseFlexibleTime(value)
		if err != nil {
			return fmt.Errorf("invalid time format: %w", err)
		}
		params.Column14 = t
	case "outfordelivery_time":
		t, err := parseFlexibleTime(value)
		if err != nil {
			return fmt.Errorf("invalid time format: %w", err)
		}
		params.Column15 = t
	case "status":
		params.Column16 = value
	default:
		return fmt.Errorf("unsupported field: %s", field)
	}

	return u.repo.UpdateShipmentDynamic(ctx, params)
}

func parseFlexibleTime(value string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format: %s", value)
}

// CountCreatedSince counts total shipments created since a given time
func (u *Usecase) CountCreatedSince(ctx context.Context, companyID uuid.UUID, since time.Time) (int64, error) {
	count, err := u.repo.CountCreatedSince(ctx, db.CountCreatedSinceParams{
		CompanyID: toNullUUID(companyID),
		CreatedAt: dbutil.ToNullTime(since),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count created: %w", err)
	}
	return count, nil
}

func (u *Usecase) CountDailyStats(ctx context.Context, companyID uuid.UUID, since time.Time) (created int64, delivered int64, err error) {
	created, err = u.repo.CountCreatedSince(ctx, db.CountCreatedSinceParams{CompanyID: toNullUUID(companyID), CreatedAt: dbutil.ToNullTime(since)})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count created: %w", err)
	}
	delivered, err = u.repo.CountDeliveredSince(ctx, db.CountDeliveredSinceParams{CompanyID: toNullUUID(companyID), UpdatedAt: dbutil.ToNullTime(since)})
	if err != nil {
		return created, 0, fmt.Errorf("failed to count delivered: %w", err)
	}
	return created, delivered, nil
}

func (u *Usecase) ListAll(ctx context.Context, companyID uuid.UUID) ([]db.Shipment, error) {
	return u.repo.ListAllShipments(ctx, toNullUUID(companyID))
}

type TransitionResult struct {
	TrackingID     string
	NewStatus      string
	UserJID        string
	RecipientEmail string
}

func (u *Usecase) ProcessTransitions(ctx context.Context, companyID uuid.UUID, now time.Time) ([]TransitionResult, error) {
	var results []TransitionResult

	transit, err := u.repo.TransitionStatusToIntransit(ctx, db.TransitionStatusToIntransitParams{CompanyID: toNullUUID(companyID), ScheduledTransitTime: dbutil.ToNullTime(now)})
	if err != nil {
		return nil, err
	}
	for _, t := range transit {
		results = append(results, TransitionResult{TrackingID: t.TrackingID, NewStatus: t.NewStatus.String, UserJID: t.UserJid, RecipientEmail: t.RecipientEmail.String})
	}

	out, err := u.repo.TransitionStatusToOutForDelivery(ctx, db.TransitionStatusToOutForDeliveryParams{CompanyID: toNullUUID(companyID), OutfordeliveryTime: dbutil.ToNullTime(now)})
	if err != nil {
		return results, err
	}
	for _, t := range out {
		results = append(results, TransitionResult{TrackingID: t.TrackingID, NewStatus: t.NewStatus.String, UserJID: t.UserJid, RecipientEmail: t.RecipientEmail.String})
	}

	delivered, err := u.repo.TransitionStatusToDelivered(ctx, db.TransitionStatusToDeliveredParams{CompanyID: toNullUUID(companyID), ExpectedDeliveryTime: dbutil.ToNullTime(now)})
	if err != nil {
		return results, err
	}
	for _, t := range delivered {
		results = append(results, TransitionResult{TrackingID: t.TrackingID, NewStatus: t.NewStatus.String, UserJID: t.UserJid, RecipientEmail: t.RecipientEmail.String})
	}

	return results, nil
}

func (u *Usecase) CountAll(ctx context.Context, companyID uuid.UUID) (int64, error) {
	return u.repo.CountShipments(ctx, toNullUUID(companyID))
}

func (u *Usecase) CountByStatus(ctx context.Context, companyID uuid.UUID) (*db.CountShipmentsByStatusRow, error) {
	stats, err := u.repo.CountShipmentsByStatus(ctx, toNullUUID(companyID))
	if err != nil {
		return nil, fmt.Errorf("failed to count by status: %w", err)
	}
	return &stats, nil
}

func (u *Usecase) RecordEvent(ctx context.Context, companyID uuid.UUID, eventType string, metadata []byte) error {
	return u.repo.RecordEvent(ctx, db.RecordEventParams{
		CompanyID: toNullUUID(companyID),
		EventType: eventType,
		Metadata:  pqtype.NullRawMessage{RawMessage: json.RawMessage(metadata), Valid: len(metadata) > 0},
	})
}

func (u *Usecase) GetTelemetryStats(ctx context.Context, companyID uuid.UUID, since time.Time) ([]db.GetTelemetryStatsRow, error) {
	return u.repo.GetTelemetryStats(ctx, db.GetTelemetryStatsParams{CompanyID: toNullUUID(companyID), CreatedAt: dbutil.ToNullTime(since)})
}

// GetRecentEvents fetches the latest telemetry events.
func (u *Usecase) GetRecentEvents(ctx context.Context, companyID uuid.UUID, limit int32) ([]db.Telemetry, error) {
	return u.repo.GetRecentEvents(ctx, db.GetRecentEventsParams{CompanyID: toNullUUID(companyID), Limit: limit})
}

// BulkUpdateStatus updates the status for multiple shipments at once.
func (u *Usecase) BulkUpdateStatus(ctx context.Context, companyID uuid.UUID, ids []string, status string) error {
	return u.repo.BulkUpdateStatus(ctx, db.BulkUpdateStatusParams{
		CompanyID: toNullUUID(companyID),
		Column2:   ids, // BulkUpdateStatusParams expects Column2
		Status:    dbutil.ToNullString(status),
	})
}

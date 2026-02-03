package localdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"webtracker-bot/internal/shipment"
)

// CreateShipment inserts a new shipment with pre-calculated timestamps.
func (c *Client) CreateShipment(ctx context.Context, s *shipment.Shipment) error {
	query := `
	INSERT INTO Shipment (
		tracking_id, user_jid, status, 
		created_at, scheduled_transit_time, outfordelivery_time, expected_delivery_time,
		sender_timezone, recipient_timezone,
		sender_name, sender_phone, origin,
		recipient_name, recipient_phone, recipient_email, recipient_id, recipient_address, destination,
		cargo_type, weight, cost
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := c.db.ExecContext(ctx, query,
		s.TrackingID, s.UserJID, s.Status,
		s.CreatedAt, s.ScheduledTransitTime, s.OutForDeliveryTime, s.ExpectedDeliveryTime,
		s.SenderTimezone, s.RecipientTimezone,
		s.SenderName, s.SenderPhone, s.Origin,
		s.RecipientName, s.RecipientPhone, s.RecipientEmail, s.RecipientID, s.RecipientAddress, s.Destination,
		s.CargoType, s.Weight, s.Cost,
	)
	return err
}

// GetShipment retrieves a shipment by its tracking ID.
func (c *Client) GetShipment(ctx context.Context, trackingID string) (*shipment.Shipment, error) {
	query := `SELECT 
		tracking_id, status, created_at, scheduled_transit_time, outfordelivery_time, expected_delivery_time,
		sender_timezone, recipient_timezone,
		sender_name, sender_phone, origin,
		recipient_name, recipient_phone, recipient_email, recipient_id, recipient_address, destination,
		cargo_type, weight, cost
		FROM Shipment WHERE tracking_id = ?`

	var s shipment.Shipment
	err := c.db.QueryRowContext(ctx, query, trackingID).Scan(
		&s.TrackingID, &s.Status, &s.CreatedAt, &s.ScheduledTransitTime, &s.OutForDeliveryTime, &s.ExpectedDeliveryTime,
		&s.SenderTimezone, &s.RecipientTimezone,
		&s.SenderName, &s.SenderPhone, &s.Origin,
		&s.RecipientName, &s.RecipientPhone, &s.RecipientEmail, &s.RecipientID, &s.RecipientAddress, &s.Destination,
		&s.CargoType, &s.Weight, &s.Cost,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetShipmentsReadyForTransit finds all PENDING shipments that have passed their transit time
func (c *Client) GetShipmentsReadyForTransit(ctx context.Context) ([]string, error) {
	query := `SELECT tracking_id FROM Shipment 
			  WHERE status = ? AND scheduled_transit_time <= ?`

	rows, err := c.db.QueryContext(ctx, query, shipment.StatusPending, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// GetShipmentsReadyForOutForDelivery finds all INTRANSIT shipments that have passed their out-for-delivery time
func (c *Client) GetShipmentsReadyForOutForDelivery(ctx context.Context) ([]string, error) {
	query := `SELECT tracking_id FROM Shipment 
			  WHERE status = ? AND outfordelivery_time <= ?`

	rows, err := c.db.QueryContext(ctx, query, shipment.StatusIntransit, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// GetShipmentsReadyForDelivery finds all OUT_FOR_DELIVERY shipments that have passed their delivery time
func (c *Client) GetShipmentsReadyForDelivery(ctx context.Context) ([]string, error) {
	query := `SELECT tracking_id FROM Shipment 
			  WHERE status = ? AND expected_delivery_time <= ?`

	rows, err := c.db.QueryContext(ctx, query, shipment.StatusOutForDelivery, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// UpdateShipmentStatus updates a single shipment's status
func (c *Client) UpdateShipmentStatus(ctx context.Context, trackingID, newStatus string) error {
	query := `UPDATE Shipment SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = ?`
	_, err := c.db.ExecContext(ctx, query, newStatus, trackingID)
	return err
}

// RunAgedCleanup enforces the data retention policy:
// 1. Delete delivered shipments older than 2 days (updated_at)
// 2. Delete ALL shipments older than 7 days (created_at)
func (c *Client) RunAgedCleanup(ctx context.Context) (int64, error) {
	twoDaysAgo := time.Now().Add(-48 * time.Hour).UTC()
	sevenDaysAgo := time.Now().Add(-168 * time.Hour).UTC()

	// 1. Cleanup Delivered (>2 days)
	res1, err := c.db.ExecContext(ctx, "DELETE FROM Shipment WHERE status = ? AND updated_at < ?", shipment.StatusDelivered, twoDaysAgo)
	if err != nil {
		return 0, fmt.Errorf("failed cleaning delivered: %w", err)
	}
	d1, _ := res1.RowsAffected()

	// 2. Cleanup All (>7 days)
	res2, err := c.db.ExecContext(ctx, "DELETE FROM Shipment WHERE created_at < ?", sevenDaysAgo)
	if err != nil {
		return d1, fmt.Errorf("failed cleaning all aged: %w", err)
	}
	d2, _ := res2.RowsAffected()

	return d1 + d2, nil
}

// CountDailyStats returns counts for Created and Delivered shipments in the last 24h
func (c *Client) CountDailyStats(ctx context.Context, since time.Time) (created int, delivered int, err error) {
	// Simple count query for created
	err = c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM Shipment WHERE created_at >= ?", since).Scan(&created)
	if err != nil {
		return 0, 0, err
	}

	// Simple count query for delivered (approximate by looking at when they *should* have been delivered,
	// or refined by adding an 'actual_delivery_time' later. For now, we trust the status update happened nearby)
	// Better: Select count where status='delivered' AND updated_at >= since (assuming we update the timestamp)
	err = c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM Shipment WHERE status = ? AND updated_at >= ?", shipment.StatusDelivered, since).Scan(&delivered)
	if err != nil {
		return created, 0, err
	}

	return created, delivered, nil
}

// GetLastTrackingByJID retrieves the most recently created tracking ID for a user.
func (c *Client) GetLastTrackingByJID(ctx context.Context, jid string) (string, error) {
	query := `SELECT tracking_id FROM Shipment WHERE user_jid = ? ORDER BY created_at DESC LIMIT 1`
	var id string
	err := c.db.QueryRowContext(ctx, query, jid).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return id, err
}

// UpdateShipmentField updates a specific field dynamically (for !edit command)
func (c *Client) UpdateShipmentField(ctx context.Context, trackingID, field, value string) error {
	// Allow-list fields to prevent SQL injection
	allowedFields := map[string]bool{
		"sender_name": true, "sender_phone": true, "origin": true,
		"recipient_name": true, "recipient_phone": true, "recipient_email": true, "recipient_id": true, "recipient_address": true, "destination": true,
		"cargo_type": true,
	}

	if !allowedFields[field] {
		return fmt.Errorf("invalid field: %s", field)
	}

	query := fmt.Sprintf("UPDATE Shipment SET %s = ?, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = ?", field)
	_, err := c.db.ExecContext(ctx, query, value, trackingID)
	return err
}

// DeleteShipment removes a shipment by ID
func (c *Client) DeleteShipment(ctx context.Context, trackingID string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM Shipment WHERE tracking_id = ?", trackingID)
	return err
}

// ListShipments returns all shipments, ordered by creation date
func (c *Client) ListShipments(ctx context.Context) ([]shipment.Shipment, error) {
	query := `SELECT 
		tracking_id, status, created_at, scheduled_transit_time, outfordelivery_time, expected_delivery_time,
		sender_timezone, recipient_timezone,
		sender_name, sender_phone, origin,
		recipient_name, recipient_phone, recipient_email, recipient_id, recipient_address, destination,
		cargo_type, weight, cost
		FROM Shipment ORDER BY created_at DESC`

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shipments []shipment.Shipment
	for rows.Next() {
		var s shipment.Shipment
		err := rows.Scan(
			&s.TrackingID, &s.Status, &s.CreatedAt, &s.ScheduledTransitTime, &s.OutForDeliveryTime, &s.ExpectedDeliveryTime,
			&s.SenderTimezone, &s.RecipientTimezone,
			&s.SenderName, &s.SenderPhone, &s.Origin,
			&s.RecipientName, &s.RecipientPhone, &s.RecipientEmail, &s.RecipientID, &s.RecipientAddress, &s.Destination,
			&s.CargoType, &s.Weight, &s.Cost,
		)
		if err == nil {
			shipments = append(shipments, s)
		}
	}
	return shipments, nil
}

// UpdateShipment updates all editable fields of a shipment
func (c *Client) UpdateShipment(ctx context.Context, s *shipment.Shipment) error {
	query := `UPDATE Shipment SET 
		status = ?, 
		sender_name = ?, sender_phone = ?, origin = ?,
		recipient_name = ?, recipient_phone = ?, recipient_email = ?, recipient_id = ?, recipient_address = ?, destination = ?,
		cargo_type = ?, weight = ?, cost = ?,
		updated_at = CURRENT_TIMESTAMP
		WHERE tracking_id = ?`

	_, err := c.db.ExecContext(ctx, query,
		s.Status,
		s.SenderName, s.SenderPhone, s.Origin,
		s.RecipientName, s.RecipientPhone, s.RecipientEmail, s.RecipientID, s.RecipientAddress, s.Destination,
		s.CargoType, s.Weight, s.Cost,
		s.TrackingID,
	)
	return err
}

// FindSimilarShipment checks for an existing shipment with the same recipient phone for this user.
func (c *Client) FindSimilarShipment(ctx context.Context, userJID, recipientPhone string) (string, error) {
	// Simple duplicate check: Same User, Same Recipient Phone
	// We ignore strict time windows as requested
	query := `SELECT tracking_id FROM Shipment 
			  WHERE user_jid = ? 
			  AND recipient_phone = ?
			  ORDER BY created_at DESC LIMIT 1`

	var id string
	err := c.db.QueryRowContext(ctx, query, userJID, recipientPhone).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return id, err
}

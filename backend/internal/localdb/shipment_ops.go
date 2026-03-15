package localdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"webtracker-bot/internal/shipment"
)

// CreateShipment inserts a new shipment. The DB trigger auto-generates:
// tracking_id, scheduled_transit_time, outfordelivery_time, expected_delivery_time.
func (c *Client) CreateShipment(ctx context.Context, s *shipment.Shipment, prefix string) (string, error) {
	if prefix == "" {
		prefix = "AWB"
	}
	// Generate tracking ID: Prefix + 9 random digits.
	// We use a more robust random factor by combining unix nano and a large random number
	seed := time.Now().UnixNano()
	randStr := fmt.Sprintf("%09d", seed%1000000000)
	trackingID := fmt.Sprintf("%s-%s", prefix, randStr)
	
	query := `
	INSERT INTO Shipment (
		tracking_id, user_jid, status,
		sender_timezone, recipient_timezone,
		sender_name, sender_phone, origin,
		recipient_name, recipient_phone, recipient_email, recipient_id, recipient_address, destination,
		cargo_type, weight, cost,
		scheduled_transit_time
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	RETURNING tracking_id`

	var returnedID string
	err := c.db.QueryRowContext(ctx, query,
		trackingID, s.UserJID, s.Status,
		s.SenderTimezone, s.RecipientTimezone,
		s.SenderName, s.SenderPhone, s.Origin,
		s.RecipientName, s.RecipientPhone, s.RecipientEmail, s.RecipientID, s.RecipientAddress, s.Destination,
		s.CargoType, s.Weight, s.Cost,
		s.ScheduledTransitTime,
	).Scan(&returnedID)
	return returnedID, err
}

// GetShipment retrieves a shipment by its tracking ID.
func (c *Client) GetShipment(ctx context.Context, trackingID string) (*shipment.Shipment, error) {
	query := `SELECT 
		tracking_id, status, created_at, scheduled_transit_time, outfordelivery_time, expected_delivery_time,
		sender_timezone, recipient_timezone,
		sender_name, sender_phone, origin,
		recipient_name, recipient_phone, recipient_email, recipient_id, recipient_address, destination,
		cargo_type, weight, cost
		FROM Shipment WHERE tracking_id = $1`

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

// ProcessStatusTransitions atomically processes all pending status changes using the DB function
func (c *Client) ProcessStatusTransitions(ctx context.Context, now time.Time) ([]struct{TrackingID, NewStatus, UserJID, RecipientEmail string}, error) {
	query := `SELECT r_tracking_id, new_status, r_user_jid, r_recipient_email FROM public.fn_process_status_transitions($1)`
	rows, err := c.db.QueryContext(ctx, query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct{TrackingID, NewStatus, UserJID, RecipientEmail string}
	for rows.Next() {
		var res struct{TrackingID, NewStatus, UserJID, RecipientEmail string}
		if err := rows.Scan(&res.TrackingID, &res.NewStatus, &res.UserJID, &res.RecipientEmail); err == nil {
			results = append(results, res)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}
	return results, nil
}

// RunAgedCleanup enforces the data retention policy using the DB function:
func (c *Client) RunAgedCleanup(ctx context.Context) (int64, error) {
	var deleted int64
	err := c.db.QueryRowContext(ctx, "SELECT deleted_count FROM public.fn_prune_aged_shipments()").Scan(&deleted)
	if err != nil {
		return 0, fmt.Errorf("failed running aged cleanup function: %w", err)
	}
	return deleted, nil
}

// CountDailyStats returns counts for Created and Delivered shipments in the last 24h
func (c *Client) CountDailyStats(ctx context.Context, since time.Time) (created int, delivered int, err error) {
	// Simple count query for created
	err = c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM Shipment WHERE created_at >= $1", since).Scan(&created)
	if err != nil {
		return 0, 0, err
	}

	// Simple count query for delivered (approximate by looking at when they *should* have been delivered,
	// or refined by adding an 'actual_delivery_time' later. For now, we trust the status update happened nearby)
	// Better: Select count where status='delivered' AND updated_at >= since (assuming we update the timestamp)
	err = c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM Shipment WHERE status = $1 AND updated_at >= $2", shipment.StatusDelivered, since).Scan(&delivered)
	if err != nil {
		return created, 0, err
	}

	return created, delivered, nil
}

// GetLastTrackingByJID retrieves the most recently created tracking ID for a user.
func (c *Client) GetLastTrackingByJID(ctx context.Context, jid string) (string, error) {
	query := `SELECT tracking_id FROM Shipment WHERE user_jid = $1 ORDER BY created_at DESC LIMIT 1`
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
		"cargo_type":             true,
		"scheduled_transit_time": true, "expected_delivery_time": true, "outfordelivery_time": true,
	}

	if !allowedFields[field] {
		return fmt.Errorf("invalid field: %s", field)
	}

	query := fmt.Sprintf("UPDATE Shipment SET %s = $1, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $2", field)
	_, err := c.db.ExecContext(ctx, query, value, trackingID)
	return err
}

// DeleteShipment removes a shipment by ID
func (c *Client) DeleteShipment(ctx context.Context, trackingID string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM Shipment WHERE tracking_id = $1", trackingID)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}
	return shipments, nil
}

// UpdateShipment updates all editable fields of a shipment
func (c *Client) UpdateShipment(ctx context.Context, s *shipment.Shipment) error {
	query := `UPDATE Shipment SET 
		status = $1, 
		sender_name = $2, sender_phone = $3, origin = $4,
		recipient_name = $5, recipient_phone = $6, recipient_email = $7, recipient_id = $8, recipient_address = $9, destination = $10,
		cargo_type = $11, weight = $12, cost = $13
		WHERE tracking_id = $14`

	_, err := c.db.ExecContext(ctx, query,
		s.Status,
		s.SenderName, s.SenderPhone, s.Origin,
		s.RecipientName, s.RecipientPhone, s.RecipientEmail, s.RecipientID, s.RecipientAddress, s.Destination,
		s.CargoType, s.Weight, s.Cost,
		s.TrackingID,
	)
	return err
}

// FindSimilarShipment checks for an existing shipment with same recipient phone, email, or ID for this user.
func (c *Client) FindSimilarShipment(ctx context.Context, userJID, phone, email, id string) (string, error) {
	// Simple duplicate check: Same User AND (Same Phone OR Same Email OR Same ID)
	query := `SELECT tracking_id FROM Shipment 
			  WHERE user_jid = $1 
			  AND (
				  (recipient_phone = $2 AND $2 != '') OR 
				  (recipient_email = $3 AND $3 != '') OR 
				  (recipient_id = $4 AND $4 != '')
			  )
			  ORDER BY created_at DESC LIMIT 1`

	var trackingID string
	err := c.db.QueryRowContext(ctx, query, userJID, phone, email, id).Scan(&trackingID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return trackingID, err
}

// GetLastShipmentIDForUser returns the tracking ID of the most recently created shipment for a given JID.
func (c *Client) GetLastShipmentIDForUser(ctx context.Context, userJID string) (string, error) {
	query := `SELECT tracking_id FROM Shipment WHERE user_jid = $1 ORDER BY created_at DESC LIMIT 1`
	var id string
	err := c.db.QueryRowContext(ctx, query, userJID).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return id, err
}

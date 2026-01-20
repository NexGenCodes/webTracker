package supabase

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/utils"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Client struct {
	db     *sql.DB
	Prefix string
}

func NewClient(dbURL string, prefix string) (*Client, error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open pgx connection: %w", err)
	}

	// Pool Limits
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{db: db, Prefix: prefix}, nil
}

// Ping checks database connectivity
func (s *Client) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.db.PingContext(ctx)
}

// withRetry handles transient network/database errors for write operations
func (s *Client) withRetry(fn func() error) error {
	var lastErr error
	for i := 0; i < 3; i++ {
		if err := fn(); err != nil {
			lastErr = err
			logger.Warn().Err(err).Int("attempt", i+1).Msg("Database write failed, retrying...")
			time.Sleep(1 * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("operation failed after 3 retries: %w", lastErr)
}

func (s *Client) CheckDuplicate(ctx context.Context, phone string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT "trackingNumber" FROM "Shipment" WHERE "receiverPhone" = $1 LIMIT 1`
	var tracking string
	err := s.db.QueryRowContext(ctx, query, phone).Scan(&tracking)
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, tracking, nil
}

func (s *Client) InsertShipment(ctx context.Context, m models.Manifest, senderPhone string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var finalID string
	err := s.withRetry(func() error {
		// Regenerate ID on every retry to prevent collisions
		randomLen := 9 - len(s.Prefix) - 1
		if randomLen < 3 {
			randomLen = 3
		}
		newID := s.Prefix + "-" + utils.GenerateShortID(randomLen)

		query := `
			INSERT INTO "Shipment" (
				"id", "trackingNumber", "status", "senderName", "senderCountry", 
				"receiverName", "receiverPhone", "receiverEmail", "receiverID", "receiverAddress", "receiverCountry", "whatsappFrom", "createdAt", "updatedAt"
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
			RETURNING "trackingNumber"`

		return s.db.QueryRowContext(ctx,
			query,
			uuid.NewString(), newID, "PENDING", m.SenderName, m.SenderCountry,
			m.ReceiverName, m.ReceiverPhone, m.ReceiverEmail, m.ReceiverID, m.ReceiverAddress, m.ReceiverCountry, senderPhone,
		).Scan(&finalID)
	})

	if err != nil {
		return "", err
	}
	return finalID, nil
}

func (s *Client) GetTodayStats(loc *time.Location) (pending, inTransit int, err error) {
	now := time.Now().In(loc)
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC()

	pendingQuery := `SELECT COUNT(*) FROM "Shipment" WHERE "status" = 'PENDING' AND "createdAt" >= $1`
	err = s.db.QueryRow(pendingQuery, midnight).Scan(&pending)
	if err != nil {
		return 0, 0, err
	}

	transitQuery := `SELECT COUNT(*) FROM "Shipment" WHERE "status" = 'IN_TRANSIT' AND "createdAt" >= $1`
	err = s.db.QueryRow(transitQuery, midnight).Scan(&inTransit)
	if err != nil {
		return 0, 0, err
	}

	return pending, inTransit, nil
}

type TransitionResult struct {
	TrackingNumber string
	WhatsappFrom   *string
}

func (s *Client) TransitionPendingToInTransit() ([]TransitionResult, error) {
	oneHourAgo := time.Now().Add(-1 * time.Hour).UTC()

	// Direct Update with Returning
	query := `
		UPDATE "Shipment" 
		SET "status" = 'IN_TRANSIT', "updatedAt" = NOW()
		WHERE "status" = 'PENDING' AND "createdAt" < $1
		RETURNING "trackingNumber", "whatsappFrom"`

	rows, err := s.db.Query(query, oneHourAgo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TransitionResult
	for rows.Next() {
		var r TransitionResult
		if err := rows.Scan(&r.TrackingNumber, &r.WhatsappFrom); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

func (s *Client) PruneStaleData() (int, error) {
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour).UTC()
	totalPruned := 0

	// Batched Pruning to avoid table locks
	for {
		query := `
			DELETE FROM "Shipment" 
			WHERE "id" IN (
				SELECT "id" FROM "Shipment" 
				WHERE "createdAt" < $1 
				LIMIT 100
			)`

		res, err := s.db.Exec(query, sevenDaysAgo)
		if err != nil {
			return totalPruned, fmt.Errorf("batch prune failed: %w", err)
		}

		rows, _ := res.RowsAffected()
		if rows == 0 {
			break
		}
		totalPruned += int(rows)
		logger.Info().Int("batch_size", int(rows)).Msg("Batch pruned records")
		time.Sleep(100 * time.Millisecond) // Pause between batches
	}

	return totalPruned, nil
}

func (s *Client) GetShipment(ctx context.Context, trackingNumber string) (*models.Shipment, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	queryFields := `
		SELECT "id", "trackingNumber", "status", "senderName", "senderCountry", 
			   "receiverName", "receiverPhone", "receiverEmail", "receiverID", "receiverAddress", "receiverCountry", 
			   "whatsappFrom", "createdAt", "updatedAt", "lastNotifiedAt"
		FROM "Shipment" WHERE "trackingNumber" = $1 LIMIT 1`

	var shipment models.Shipment
	err := s.db.QueryRowContext(ctx, queryFields, trackingNumber).Scan(
		&shipment.ID, &shipment.TrackingNumber, &shipment.Status, &shipment.SenderName, &shipment.SenderCountry,
		&shipment.ReceiverName, &shipment.ReceiverPhone, &shipment.ReceiverEmail, &shipment.ReceiverID, &shipment.ReceiverAddress, &shipment.ReceiverCountry,
		&shipment.WhatsappFrom, &shipment.CreatedAt, &shipment.UpdatedAt, &shipment.LastNotifiedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &shipment, nil
}

func (s *Client) UpdateShipmentField(ctx context.Context, trackingID string, field string, value string) error {
	// Whitelist allowed fields to prevent arbitrary column updates
	allowed := map[string]string{
		// Receiver Fields
		"name":            "receiverName",
		"receivername":    "receiverName",
		"receiver":        "receiverName",
		"phone":           "receiverPhone",
		"receiverphone":   "receiverPhone",
		"mobile":          "receiverPhone",
		"address":         "receiverAddress",
		"receiveraddress": "receiverAddress",
		"addr":            "receiverAddress",
		"country":         "receiverCountry",
		"receivercountry": "receiverCountry",
		"destination":     "receiverCountry",
		"email":           "receiverEmail",
		"receiveremail":   "receiverEmail",
		"id":              "receiverID",
		"receiverid":      "receiverID",

		// Sender Fields
		"sender":        "senderName",
		"sendername":    "senderName",
		"origin":        "senderCountry",
		"sendercountry": "senderCountry",
		"from":          "senderCountry",
	}

	dbField, ok := allowed[strings.ToLower(field)]
	if !ok {
		return fmt.Errorf("field %s not allowed for editing", field)
	}

	query := fmt.Sprintf(`UPDATE "Shipment" SET "%s" = $1, "updatedAt" = NOW() WHERE "trackingNumber" = $2`, dbField)
	_, err := s.db.ExecContext(ctx, query, value, trackingID)
	return err
}

func (s *Client) GetLastTrackingByJID(ctx context.Context, jid string) (string, error) {
	query := `SELECT "trackingNumber" FROM "Shipment" WHERE "whatsappFrom" = $1 ORDER BY "createdAt" DESC LIMIT 1`
	var tracking string
	err := s.db.QueryRowContext(ctx, query, jid).Scan(&tracking)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return tracking, err
}

func (s *Client) GetPendingNotifications() ([]models.NotificationJob, error) {
	query := `
		SELECT "trackingNumber", "status", "whatsappFrom"
		FROM "Shipment"
		WHERE ("lastNotifiedAt" IS NULL OR "updatedAt" > "lastNotifiedAt")
		AND "status" = 'IN_TRANSIT'
		AND "whatsappFrom" IS NOT NULL
		LIMIT 50`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.NotificationJob
	for rows.Next() {
		var j models.NotificationJob
		if err := rows.Scan(&j.TrackingNumber, &j.Status, &j.WhatsappFrom); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (s *Client) MarkAsNotified(tracking string) error {
	query := `UPDATE "Shipment" SET "lastNotifiedAt" = NOW() WHERE "trackingNumber" = $1`
	_, err := s.db.Exec(query, tracking)
	return err
}

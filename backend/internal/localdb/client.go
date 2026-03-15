package localdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"webtracker-bot/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Client struct {
	db *sql.DB
}

func NewClient(dsn string) (*Client, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	// Connection pool settings suitable for 1GB RAM VPS
	db.SetMaxOpenConns(5)    // Small pool
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(1 * time.Hour)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping session database: %w", err)
	}

	client := &Client{db: db}
	if err := client.InitSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	logger.Info().Msg("Local DB (PostgreSQL) initialized")
	return client, nil
}

func (c *Client) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS public.SystemConfig (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS public.UserPreference (
		jid TEXT PRIMARY KEY,
		language TEXT NOT NULL DEFAULT 'en',
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS public.GroupAuthority (
		jid TEXT PRIMARY KEY,
		is_authorized BOOLEAN NOT NULL DEFAULT FALSE,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS public.Shipment (
		tracking_id TEXT PRIMARY KEY,
		user_jid TEXT NOT NULL,
		status TEXT DEFAULT 'pending',
		
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		scheduled_transit_time TIMESTAMP,
		outfordelivery_time TIMESTAMP,
		expected_delivery_time TIMESTAMP,
		
		sender_timezone TEXT,
		recipient_timezone TEXT,

		sender_name TEXT,
		sender_phone TEXT,
		origin TEXT,
		recipient_name TEXT,
		recipient_phone TEXT,
		recipient_email TEXT,
		recipient_id TEXT,
		recipient_address TEXT,
		destination TEXT,
		cargo_type TEXT,
		weight DOUBLE PRECISION,
		cost DOUBLE PRECISION,
		
		-- Metadata for easy deletes or bulk ops
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_shipment_triggers_pending ON public.Shipment(status, scheduled_transit_time) WHERE status = 'pending';
	CREATE INDEX IF NOT EXISTS idx_shipment_triggers_transit ON public.Shipment(status, outfordelivery_time) WHERE status = 'intransit';
	CREATE INDEX IF NOT EXISTS idx_shipment_triggers_outfordelivery ON public.Shipment(status, expected_delivery_time) WHERE status = 'outfordelivery';
	CREATE INDEX IF NOT EXISTS idx_shipment_user_jid ON public.Shipment(user_jid);

	-- Uniqueness Constraints (Recipient Details)
	-- We use partial indices to allow multiple NULLs but enforce uniqueness on values
	CREATE UNIQUE INDEX IF NOT EXISTS idx_shipment_unique_phone ON public.Shipment(recipient_phone) WHERE recipient_phone IS NOT NULL AND recipient_phone != '';
	CREATE UNIQUE INDEX IF NOT EXISTS idx_shipment_unique_email ON public.Shipment(recipient_email) WHERE recipient_email IS NOT NULL AND recipient_email != '';
	CREATE UNIQUE INDEX IF NOT EXISTS idx_shipment_unique_id ON public.Shipment(recipient_id) WHERE recipient_id IS NOT NULL AND recipient_id != '';
	`
	_, err := c.db.ExecContext(ctx, query)
	return err
}

// ResetDB drops all tables and re-initializes the schema.
func (c *Client) ResetDB(ctx context.Context) error {
	query := `
	DROP TABLE IF EXISTS public.Shipment CASCADE;
	DROP TABLE IF EXISTS public.UserPreference CASCADE;
	DROP TABLE IF EXISTS public.GroupAuthority CASCADE;
	DROP TABLE IF EXISTS public.SystemConfig CASCADE;
	`
	if _, err := c.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}
	return c.InitSchema(ctx)
}

// GetSystemConfig fetches a persistent configuration value.
func (c *Client) GetSystemConfig(ctx context.Context, key string) (string, error) {
	query := `SELECT value FROM SystemConfig WHERE key = $1`
	var val string
	err := c.db.QueryRowContext(ctx, query, key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

// SetSystemConfig updates or inserts a persistent configuration value.
func (c *Client) SetSystemConfig(ctx context.Context, key string, value string) error {
	query := `
	INSERT INTO SystemConfig (key, value, updated_at) 
	VALUES ($1, $2, CURRENT_TIMESTAMP)
	ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP;
	`
	_, err := c.db.ExecContext(ctx, query, key, value)
	return err
}

func (c *Client) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

func (c *Client) Close() error {
	return c.db.Close()
}

// GetUserLanguage fetches the preferred language for a JID. Defaults to "en" if not found.
func (c *Client) GetUserLanguage(ctx context.Context, jid string) (string, error) {
	query := `SELECT language FROM UserPreference WHERE jid = $1`
	var lang string
	err := c.db.QueryRowContext(ctx, query, jid).Scan(&lang)
	if err == sql.ErrNoRows {
		return "en", nil
	}
	if err != nil {
		return "en", err
	}
	return lang, nil
}

// SetUserLanguage updates or inserts the preferred language for a JID.
func (c *Client) SetUserLanguage(ctx context.Context, jid string, lang string) error {
	query := `
	INSERT INTO UserPreference (jid, language, updated_at) 
	VALUES ($1, $2, CURRENT_TIMESTAMP)
	ON CONFLICT(jid) DO UPDATE SET language = EXCLUDED.language, updated_at = CURRENT_TIMESTAMP;
	`
	_, err := c.db.ExecContext(ctx, query, jid, lang)
	return err
}

// GetGroupAuthority checks the cache for bot authorization in a group.
// Returns (isAuthorized, exists, error)
func (c *Client) GetGroupAuthority(ctx context.Context, jid string) (bool, bool, error) {
	query := `SELECT is_authorized, updated_at FROM GroupAuthority WHERE jid = $1`
	var isAuthorized bool
	var updatedAt time.Time
	err := c.db.QueryRowContext(ctx, query, jid).Scan(&isAuthorized, &updatedAt)
	if err == sql.ErrNoRows {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}

	return isAuthorized, true, nil
}

// SetGroupAuthority updates the authority cache for a group.
func (c *Client) SetGroupAuthority(ctx context.Context, jid string, isAuthorized bool) error {
	query := `
	INSERT INTO GroupAuthority (jid, is_authorized, updated_at) 
	VALUES ($1, $2, CURRENT_TIMESTAMP)
	ON CONFLICT(jid) DO UPDATE SET is_authorized = EXCLUDED.is_authorized, updated_at = CURRENT_TIMESTAMP;
	`
	_, err := c.db.ExecContext(ctx, query, jid, isAuthorized)
	return err
}

// HasAuthorizedGroups checks if the bot is admin in at least one cached group.
func (c *Client) HasAuthorizedGroups(ctx context.Context) (bool, error) {
	query := `SELECT COUNT(*) FROM GroupAuthority WHERE is_authorized = TRUE`
	var count int
	err := c.db.QueryRowContext(ctx, query).Scan(&count)
	return count > 0, err
}

// GetAuthorizedGroups returns a list of all group JIDs where the bot is authorized.
func (c *Client) GetAuthorizedGroups(ctx context.Context) ([]string, error) {
	query := `SELECT jid FROM GroupAuthority WHERE is_authorized = TRUE`
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []string
	for rows.Next() {
		var jid string
		if err := rows.Scan(&jid); err == nil {
			groups = append(groups, jid)
		}
	}
	return groups, nil
}

// CountAuthorizedGroups returns the total number of authorized groups.
func (c *Client) CountAuthorizedGroups(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM GroupAuthority WHERE is_authorized = TRUE`
	var count int
	err := c.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

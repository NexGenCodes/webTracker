package localdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"webtracker-bot/internal/logger"

	_ "modernc.org/sqlite"
)

type Client struct {
	db *sql.DB
}

func NewClient(dbPath string) (*Client, error) {
	// Use same WAL settings as whatsmeow to ensure compatibility and performance
	// _pragma=busy_timeout(5000) is crucial for concurrent access
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite connection: %w", err)
	}

	// Connection pool settings suitable for SQLite
	db.SetMaxOpenConns(1) // SQLite writes are serialized anyway; keeping 1 connection avoids lock contention
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // Reuse forever

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping session database: %w", err)
	}

	client := &Client{db: db}
	if err := client.initSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	logger.Info().Str("path", dbPath).Msg("Local DB (SQLite) initialized")
	return client, nil
}

func (c *Client) initSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS UserPreference (
		jid TEXT PRIMARY KEY,
		language TEXT NOT NULL DEFAULT 'en',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS GroupAuthority (
		jid TEXT PRIMARY KEY,
		is_authorized BOOLEAN NOT NULL DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := c.db.ExecContext(ctx, query)
	return err
}

func (c *Client) Close() error {
	return c.db.Close()
}

// GetUserLanguage fetches the preferred language for a JID. Defaults to "en" if not found.
func (c *Client) GetUserLanguage(ctx context.Context, jid string) (string, error) {
	query := `SELECT language FROM UserPreference WHERE jid = ?`
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
	VALUES (?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(jid) DO UPDATE SET language = excluded.language, updated_at = CURRENT_TIMESTAMP;
	`
	_, err := c.db.ExecContext(ctx, query, jid, lang)
	return err
}

// GetGroupAuthority checks the cache for bot authorization in a group.
// Returns (isAuthorized, exists, error)
func (c *Client) GetGroupAuthority(ctx context.Context, jid string) (bool, bool, error) {
	query := `SELECT is_authorized, updated_at FROM GroupAuthority WHERE jid = ?`
	var isAuthorized bool
	var updatedAt time.Time
	err := c.db.QueryRowContext(ctx, query, jid).Scan(&isAuthorized, &updatedAt)
	if err == sql.ErrNoRows {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}

	// Cache TTL: 1 hour
	if time.Since(updatedAt) > time.Hour {
		return false, false, nil // Treat as "not in cache"
	}

	return isAuthorized, true, nil
}

// SetGroupAuthority updates the authority cache for a group.
func (c *Client) SetGroupAuthority(ctx context.Context, jid string, isAuthorized bool) error {
	query := `
	INSERT INTO GroupAuthority (jid, is_authorized, updated_at) 
	VALUES (?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(jid) DO UPDATE SET is_authorized = excluded.is_authorized, updated_at = CURRENT_TIMESTAMP;
	`
	_, err := c.db.ExecContext(ctx, query, jid, isAuthorized)
	return err
}

// HasAuthorizedGroups checks if the bot is admin in at least one cached group.
func (c *Client) HasAuthorizedGroups(ctx context.Context) (bool, error) {
	query := `SELECT COUNT(*) FROM GroupAuthority WHERE is_authorized = 1 AND updated_at > ?`
	var count int
	// Use same TTL as GetGroupAuthority (1 hour)
	threshold := time.Now().Add(-time.Hour)
	err := c.db.QueryRowContext(ctx, query, threshold).Scan(&count)
	return count > 0, err
}

// GetAuthorizedGroups returns a list of all group JIDs where the bot is authorized.
func (c *Client) GetAuthorizedGroups(ctx context.Context) ([]string, error) {
	query := `SELECT jid FROM GroupAuthority WHERE is_authorized = 1 AND updated_at > ?`
	threshold := time.Now().Add(-time.Hour)
	rows, err := c.db.QueryContext(ctx, query, threshold)
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

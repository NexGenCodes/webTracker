package dbconn

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"webtracker-bot/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connect database connection string
func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open pgx connection: %w", err)
	}

	// Minimal pool sizes (serverless or VPS safe)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(1 * time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping session database: %w", err)
	}

	logger.Info().Msg("Database connected (SQLC Pgx pool)")
	return db, nil
}

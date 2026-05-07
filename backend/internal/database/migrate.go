package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	web_sql "webtracker-bot/sql"
	"webtracker-bot/internal/logger"
)

// RunMigrations runs all embedded database migrations in an idempotent manner.
func RunMigrations(ctx context.Context, db *sql.DB) error {
	// 1. Ensure migrations table exists
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// 2. Read embedded migrations
	entries, err := web_sql.MigrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read embedded migrations: %w", err)
	}

	var sqlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() {
			sqlFiles = append(sqlFiles, entry.Name())
		}
	}
	sort.Strings(sqlFiles)

	if len(sqlFiles) == 0 {
		return nil
	}

	// 3. Apply migrations
	for _, fileName := range sqlFiles {
		var exists bool
		err = db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", fileName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", fileName, err)
		}

		if exists {
			continue // Already applied
		}

		logger.Info().Str("migration", fileName).Msg("Applying migration...")
		
		content, err := web_sql.MigrationsFS.ReadFile("migrations/" + fileName)
		if err != nil {
			return fmt.Errorf("failed to read embedded sql file %s: %w", fileName, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", fileName, err)
		}

		_, err = tx.ExecContext(ctx, string(content))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", fileName, err)
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", fileName)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", fileName, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", fileName, err)
		}
		
		logger.Info().Str("migration", fileName).Msg("Successfully applied migration")
	}

	return nil
}

package database

import (
	"context"
	"database/sql"
	"fmt"

	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/logger"

	"golang.org/x/crypto/bcrypt"
)

// SeedSuperAdmin creates the initial super admin in the DB from env vars.
// Idempotent — skips if the email already exists.
func SeedSuperAdmin(ctx context.Context, sqlPool *sql.DB, email, password string) error {
	if email == "" || password == "" {
		logger.Warn().Msg("SeedSuperAdmin: SUPERADMIN_COMPANY_EMAIL or SUPERADMIN_PASSWORD not set, skipping")
		return nil
	}

	querier := db.New(sqlPool)

	_, err := querier.GetSuperAdminByEmail(ctx, email)
	if err == nil {
		logger.Info().Str("email", email).Msg("SeedSuperAdmin: super admin already exists, skipping")
		return nil
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("SeedSuperAdmin: failed to check existing super admin: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("SeedSuperAdmin: failed to hash password: %w", err)
	}

	_, err = querier.CreateSuperAdmin(ctx, db.CreateSuperAdminParams{
		Email:        email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return fmt.Errorf("SeedSuperAdmin: failed to create super admin: %w", err)
	}

	logger.Info().Str("email", email).Msg("SeedSuperAdmin: super admin created from env")
	return nil
}

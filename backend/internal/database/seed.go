package database

import (
	"context"
	"database/sql"
	"fmt"

	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/logger"

	"github.com/google/uuid"
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

	// Ensure a company exists for the super admin (idempotent)
	if err := seedSuperAdminCompany(ctx, querier, email); err != nil {
		return fmt.Errorf("SeedSuperAdmin: failed to seed company: %w", err)
	}

	return nil
}

// seedSuperAdminCompany ensures the super admin has a default company.
// Idempotent — skips if the company already exists for this email.
func seedSuperAdminCompany(ctx context.Context, querier *db.Queries, email string) error {
	existing, err := querier.GetCompanyByEmail(ctx, email)
	if err == nil {
		// Company exists — ensure plan is unlimited
		if existing.PlanType.String != "unlimited" {
			if err := querier.UpdateCompanyPlan(ctx, db.UpdateCompanyPlanParams{
				ID:       existing.ID,
				PlanType: sql.NullString{String: "unlimited", Valid: true},
			}); err != nil {
				return fmt.Errorf("failed to update plan to unlimited: %w", err)
			}
			logger.Info().Str("email", email).Msg("SeedSuperAdmin: updated company plan to unlimited")
		}
		return nil
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing company: %w", err)
	}

	// Create default Airwaybill company for the super admin
	company, err := querier.CreateCompany(ctx, db.CreateCompanyParams{
		AdminEmail: email,
		Name:       sql.NullString{String: "Airwaybill", Valid: true},
		SetupToken: sql.NullString{String: uuid.New().String(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create company: %w", err)
	}

	// Set unlimited plan
	if err := querier.UpdateCompanyPlan(ctx, db.UpdateCompanyPlanParams{
		ID:       company.ID,
		PlanType: sql.NullString{String: "unlimited", Valid: true},
	}); err != nil {
		return fmt.Errorf("failed to set unlimited plan: %w", err)
	}

	// Set auth status to active (no OTP/linking needed)
	if err := querier.UpdateCompanyAuthStatus(ctx, db.UpdateCompanyAuthStatusParams{
		ID:         company.ID,
		AuthStatus: sql.NullString{String: "active", Valid: true},
	}); err != nil {
		return fmt.Errorf("failed to set auth status: %w", err)
	}

	logger.Info().Str("email", email).Str("company_id", company.ID.String()).Msg("SeedSuperAdmin: Airwaybill company created")
	return nil
}

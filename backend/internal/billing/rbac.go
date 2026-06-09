package billing

import (
	"context"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
)

// SuperAdminChecker is a minimal interface for looking up super admin records.
// This avoids importing the full db package and prevents circular dependencies.
type SuperAdminChecker interface {
	GetSuperAdminByEmail(ctx context.Context, email string) (struct{ Email string }, error)
}

// RoleSuperAdmin is the canonical role string for super admin users.
const RoleSuperAdmin = "super_admin"

// IsSuperAdminRole checks the JWT role claim.
// Use this in API handlers where the authenticated user's JWT claims are available.
func IsSuperAdminRole(role string) bool {
	return role == RoleSuperAdmin
}

// IsSuperAdminByEmail checks whether the given email belongs to a super admin
// using the database as the primary source of truth, with the env var as a fallback.
// Use this in background workers and cron jobs where there is no JWT context.
func IsSuperAdminByEmail(ctx context.Context, cfg *config.Config, email string, dbCheck func(ctx context.Context, email string) bool) bool {
	// 1. Database check (primary source of truth)
	if dbCheck != nil && dbCheck(ctx, email) {
		return true
	}

	// 2. Env var fallback (backward compatibility)
	if cfg != nil && cfg.SuperAdminCompanyEmail != "" && email == cfg.SuperAdminCompanyEmail {
		logger.Debug().Str("email", email).Msg("[RBAC] Super admin matched via env var fallback (deprecated)")
		return true
	}

	return false
}

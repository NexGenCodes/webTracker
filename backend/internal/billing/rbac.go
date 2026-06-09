package billing

import (
	"context"

	"webtracker-bot/internal/config"
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
// using the env var as the single source of truth.
// Use this in background workers and cron jobs where there is no JWT context.
func IsSuperAdminByEmail(cfg *config.Config, email string) bool {
	if cfg != nil && cfg.SuperAdminCompanyEmail != "" && email == cfg.SuperAdminCompanyEmail {
		return true
	}
	return false
}

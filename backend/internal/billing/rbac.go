package billing

import (
	"context"
)

// SuperAdminChecker is a minimal interface for looking up super admin records.
// This avoids importing the full db package and prevents circular dependencies.
type SuperAdminChecker interface {
	IsSuperAdmin(ctx context.Context, email string) (bool, error)
}

// RoleSuperAdmin is the canonical role string for super admin users.
const RoleSuperAdmin = "super_admin"

// IsSuperAdminRole checks the JWT role claim.
// Use this in API handlers where the authenticated user's JWT claims are available.
func IsSuperAdminRole(role string) bool {
	return role == RoleSuperAdmin
}

// IsSuperAdminByEmail checks whether the given email belongs to a super admin
// using the provided checker interface (backed by DB or env).
// Workers and cron jobs should use this with a DB-backed checker.
func IsSuperAdminByEmail(ctx context.Context, checker SuperAdminChecker, email string) (bool, error) {
	if checker == nil {
		return false, nil
	}
	return checker.IsSuperAdmin(ctx, email)
}

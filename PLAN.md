# Plan: Fix VPS Bot Crash

## Impact Analysis
- Modifies `backend/internal/localdb/client.go` to explicitly declare `public.` prefix for index creations to avoid Neon DB connection pool schema visibility errors.
- Committing to `main` will trigger the GitHub workflows to recompile and deploy the backend.

## Breaking Changes
- None. This is a stability hotfix.

## Verification
- SSH into the target server (`13.60.207.149`) after the CI/CD pipeline completes and verify `systemctl status webtracker-bot` is active and healthy.

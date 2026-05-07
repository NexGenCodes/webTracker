package sql

import "embed"

//go:embed migrations/*.sql
// MigrationsFS holds the embedded SQL migration scripts.
var MigrationsFS embed.FS

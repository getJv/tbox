package migrations

import (
	"embed"
)

// FS holds the embedded SQL migration files for the qworkers package.
// These migrations can be executed using goose or via the RunMigration helper.
//
//go:embed *.sql
var FS embed.FS

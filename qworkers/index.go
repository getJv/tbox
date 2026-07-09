package qworkers

import (
	"database/sql"
	"fmt"

	"github.com/getjv/tbox/qworkers/migrations"
	"github.com/pressly/goose/v3"
)

// SchemaProvider is an implementation of the MigrationProvider interface.
type SchemaProvider struct{}

// EnsureSchema ensures that the necessary tables for the worker exist in the database.
func (p *SchemaProvider) EnsureSchema(db *sql.DB) error {
	return RunMigration(db, "up")
}

func (p *SchemaProvider) Name() string {
	return "qWorkers"
}

// RunMigration allows executing specific migration commands (up, down, status, etc.) for qworkers.
func RunMigration(db *sql.DB, command string, args ...string) error {
	goose.SetBaseFS(migrations.FS)

	// Isolated configurations for the qworkers package
	prevTable := goose.TableName()
	goose.SetTableName("qworkers_db_version")
	defer goose.SetTableName(prevTable)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	// Executes the requested command on the package migrations
	if err := goose.Run(command, db, ".", args...); err != nil {
		return fmt.Errorf("qworkers migration (%s): %w", command, err)
	}

	return nil
}

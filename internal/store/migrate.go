package store

import (
	"database/sql"
	"embed"
	"io/fs"
)

//go:embed schema.sql
var schemaFS embed.FS

// Migrate runs the schema migration.
func Migrate(db *sql.DB) error {
	sqlBytes, err := fs.ReadFile(schemaFS, "schema.sql")
	if err != nil {
		return err
	}
	if _, err := db.Exec(string(sqlBytes)); err != nil {
		return err
	}
	// Add workspace color column for DBs created before this field existed.
	_, _ = db.Exec("ALTER TABLE workspaces ADD COLUMN color TEXT")
	// Add project color column for DBs created before this field existed.
	_, _ = db.Exec("ALTER TABLE projects ADD COLUMN color TEXT")
	return nil
}

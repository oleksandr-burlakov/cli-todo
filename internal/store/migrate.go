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
	_, err = db.Exec(string(sqlBytes))
	return err
}

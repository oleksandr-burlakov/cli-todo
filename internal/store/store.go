package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DBPath returns the default path for the SQLite database (cross-platform).
func DBPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	appDir := filepath.Join(dir, "cli-todo")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}
	return filepath.Join(appDir, "todo.db"), nil
}

// Open opens the SQLite database and runs migrations.
func Open(path string) (*sql.DB, error) {
	if path == "" {
		var err error
		path, err = DBPath()
		if err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := Migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

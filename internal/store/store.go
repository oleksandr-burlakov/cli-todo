package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DBPath returns the default path for the SQLite database.
// Uses "todo.db" next to the executable so the file is in a predictable place
// whether you run from a terminal (cd folder; ./todo) or by double-clicking the exe.
func DBPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		// Fallback to current directory if we can't get executable path
		dir, err2 := os.Getwd()
		if err2 != nil {
			return "", fmt.Errorf("executable: %w", err)
		}
		return filepath.Join(dir, "todo.db"), nil
	}
	dir := filepath.Dir(exe)
	return filepath.Join(dir, "todo.db"), nil
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

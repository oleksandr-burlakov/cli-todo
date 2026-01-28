package store

import (
	"database/sql"

	"github.com/cli-todo/internal/models"
)

func CreateWorkspace(db *sql.DB, name string) (models.Workspace, error) {
	res, err := db.Exec("INSERT INTO workspaces (name) VALUES (?)", name)
	if err != nil {
		return models.Workspace{}, err
	}
	id, _ := res.LastInsertId()
	return GetWorkspace(db, id)
}

func GetWorkspace(db *sql.DB, id int64) (models.Workspace, error) {
	var w models.Workspace
	err := db.QueryRow("SELECT id, name, created_at FROM workspaces WHERE id = ?", id).
		Scan(&w.ID, &w.Name, &w.CreatedAt)
	return w, err
}

func GetWorkspaceByName(db *sql.DB, name string) (models.Workspace, error) {
	var w models.Workspace
	err := db.QueryRow("SELECT id, name, created_at FROM workspaces WHERE name = ?", name).
		Scan(&w.ID, &w.Name, &w.CreatedAt)
	return w, err
}

func ListWorkspaces(db *sql.DB) ([]models.Workspace, error) {
	rows, err := db.Query("SELECT id, name, created_at FROM workspaces ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Workspace
	for rows.Next() {
		var w models.Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, w)
	}
	return list, rows.Err()
}

func DeleteWorkspace(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM workspaces WHERE id = ?", id)
	return err
}

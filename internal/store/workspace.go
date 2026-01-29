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
	var color sql.NullString
	err := db.QueryRow("SELECT id, name, color, created_at FROM workspaces WHERE id = ?", id).
		Scan(&w.ID, &w.Name, &color, &w.CreatedAt)
	if err != nil {
		return models.Workspace{}, err
	}
	if color.Valid {
		w.Color = color.String
	}
	return w, nil
}

func GetWorkspaceByName(db *sql.DB, name string) (models.Workspace, error) {
	var w models.Workspace
	var color sql.NullString
	err := db.QueryRow("SELECT id, name, color, created_at FROM workspaces WHERE name = ?", name).
		Scan(&w.ID, &w.Name, &color, &w.CreatedAt)
	if err != nil {
		return models.Workspace{}, err
	}
	if color.Valid {
		w.Color = color.String
	}
	return w, nil
}

func ListWorkspaces(db *sql.DB) ([]models.Workspace, error) {
	rows, err := db.Query("SELECT id, name, color, created_at FROM workspaces ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Workspace
	for rows.Next() {
		var w models.Workspace
		var color sql.NullString
		if err := rows.Scan(&w.ID, &w.Name, &color, &w.CreatedAt); err != nil {
			return nil, err
		}
		if color.Valid {
			w.Color = color.String
		}
		list = append(list, w)
	}
	return list, rows.Err()
}

func UpdateWorkspace(db *sql.DB, id int64, name string) (models.Workspace, error) {
	_, err := db.Exec("UPDATE workspaces SET name = ? WHERE id = ?", name, id)
	if err != nil {
		return models.Workspace{}, err
	}
	return GetWorkspace(db, id)
}

func SetWorkspaceColor(db *sql.DB, id int64, color string) (models.Workspace, error) {
	var val interface{} = nil
	if color != "" {
		val = color
	}
	_, err := db.Exec("UPDATE workspaces SET color = ? WHERE id = ?", val, id)
	if err != nil {
		return models.Workspace{}, err
	}
	return GetWorkspace(db, id)
}

func DeleteWorkspace(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM workspaces WHERE id = ?", id)
	return err
}

package store

import (
	"database/sql"

	"github.com/cli-todo/internal/models"
)

func CreateProject(db *sql.DB, workspaceID int64, name string) (models.Project, error) {
	res, err := db.Exec("INSERT INTO projects (workspace_id, name) VALUES (?, ?)", workspaceID, name)
	if err != nil {
		return models.Project{}, err
	}
	id, _ := res.LastInsertId()
	return GetProject(db, id)
}

func GetProject(db *sql.DB, id int64) (models.Project, error) {
	var p models.Project
	var color sql.NullString
	err := db.QueryRow("SELECT id, workspace_id, name, color, created_at FROM projects WHERE id = ?", id).
		Scan(&p.ID, &p.WorkspaceID, &p.Name, &color, &p.CreatedAt)
	if err != nil {
		return p, err
	}
	if color.Valid {
		p.Color = color.String
	}
	return p, nil
}

func ListProjects(db *sql.DB, workspaceID int64) ([]models.Project, error) {
	rows, err := db.Query("SELECT id, workspace_id, name, color, created_at FROM projects WHERE workspace_id = ? ORDER BY name", workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Project
	for rows.Next() {
		var p models.Project
		var color sql.NullString
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &color, &p.CreatedAt); err != nil {
			return nil, err
		}
		if color.Valid {
			p.Color = color.String
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func UpdateProject(db *sql.DB, id int64, name string) (models.Project, error) {
	_, err := db.Exec("UPDATE projects SET name = ? WHERE id = ?", name, id)
	if err != nil {
		return models.Project{}, err
	}
	return GetProject(db, id)
}

func SetProjectColor(db *sql.DB, id int64, color string) (models.Project, error) {
	var val interface{}
	if color != "" {
		val = color
	}
	_, err := db.Exec("UPDATE projects SET color = ? WHERE id = ?", val, id)
	if err != nil {
		return models.Project{}, err
	}
	return GetProject(db, id)
}

func DeleteProject(db *sql.DB, id int64) error {
	if _, err := db.Exec("UPDATE tasks SET project_id = NULL WHERE project_id = ?", id); err != nil {
		return err
	}
	_, err := db.Exec("DELETE FROM projects WHERE id = ?", id)
	return err
}

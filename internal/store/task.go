package store

import (
	"database/sql"
	"time"

	"github.com/cli-todo/internal/models"
)

func CreateTask(db *sql.DB, workspaceID int64, projectID *int64, title, description, status, priority string, dueDate *time.Time) (models.Task, error) {
	if status == "" {
		status = "todo"
	}
	res, err := db.Exec(
		`INSERT INTO tasks (workspace_id, project_id, title, description, status, priority, due_date) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		workspaceID, projectID, title, description, status, priority, nullTime(dueDate),
	)
	if err != nil {
		return models.Task{}, err
	}
	id, _ := res.LastInsertId()
	return GetTask(db, id)
}

func GetTask(db *sql.DB, id int64) (models.Task, error) {
	var t models.Task
	var desc, pri sql.NullString
	var projID sql.NullInt64
	var due sql.NullTime
	err := db.QueryRow(
		"SELECT id, workspace_id, project_id, title, description, status, priority, due_date, created_at, updated_at FROM tasks WHERE id = ?", id,
	).Scan(&t.ID, &t.WorkspaceID, &projID, &t.Title, &desc, &t.Status, &pri, &due, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return models.Task{}, err
	}
	if projID.Valid {
		t.ProjectID = &projID.Int64
	}
	t.Description = desc.String
	t.Priority = pri.String
	if due.Valid {
		t.DueDate = &due.Time
	}
	return t, nil
}

func ListTasks(db *sql.DB, workspaceID int64, projectID *int64) ([]models.Task, error) {
	var rows *sql.Rows
	var err error
	if projectID == nil {
		rows, err = db.Query(
			`SELECT id, workspace_id, project_id, title, description, status, priority, due_date, created_at, updated_at FROM tasks WHERE workspace_id = ? AND project_id IS NULL ORDER BY created_at`,
			workspaceID,
		)
	} else {
		rows, err = db.Query(
			`SELECT id, workspace_id, project_id, title, description, status, priority, due_date, created_at, updated_at FROM tasks WHERE workspace_id = ? AND project_id = ? ORDER BY created_at`,
			workspaceID, *projectID,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func ListAllTasksInWorkspace(db *sql.DB, workspaceID int64) ([]models.Task, error) {
	rows, err := db.Query(
		`SELECT id, workspace_id, project_id, title, description, status, priority, due_date, created_at, updated_at FROM tasks WHERE workspace_id = ? ORDER BY project_id, created_at`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func UpdateTask(db *sql.DB, id int64, title, description, status, priority string, dueDate *time.Time) (models.Task, error) {
	_, err := db.Exec(
		`UPDATE tasks SET title = ?, description = ?, status = ?, priority = ?, due_date = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		title, description, status, priority, nullTime(dueDate), id,
	)
	if err != nil {
		return models.Task{}, err
	}
	return GetTask(db, id)
}

func DeleteTask(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func scanTasks(rows *sql.Rows) ([]models.Task, error) {
	var list []models.Task
	for rows.Next() {
		var t models.Task
		var desc, pri sql.NullString
		var projID sql.NullInt64
		var due sql.NullTime
		if err := rows.Scan(&t.ID, &t.WorkspaceID, &projID, &t.Title, &desc, &t.Status, &pri, &due, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		if projID.Valid {
			t.ProjectID = &projID.Int64
		}
		t.Description = desc.String
		t.Priority = pri.String
		if due.Valid {
			t.DueDate = &due.Time
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

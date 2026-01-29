package models

import "time"

type Workspace struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color,omitempty"` // e.g. "green", "blue", "#ff0000"
	CreatedAt time.Time `json:"created_at"`
}

type Project struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	Name        string    `json:"name"`
	Color       string    `json:"color,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Task struct {
	ID          int64      `json:"id"`
	WorkspaceID int64      `json:"workspace_id"`
	ProjectID   *int64     `json:"project_id,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"` // todo, in_progress, done
	Priority    string     `json:"priority,omitempty"` // low, medium, high
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

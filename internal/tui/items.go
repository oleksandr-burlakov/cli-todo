package tui

import "github.com/cli-todo/internal/models"

// list.Item + Title/Description for bubbles list
type workspaceItem struct {
	models.Workspace
}

func (w workspaceItem) Title() string       { return w.Name }
func (w workspaceItem) Description() string  { return "" }
func (w workspaceItem) FilterValue() string   { return w.Name }

type projectItem struct {
	ID          *int64
	Name        string
	IsDefault   bool
	Color       string
}

func (p projectItem) Title() string       { return p.Name }
func (p projectItem) Description() string { return "" }
func (p projectItem) FilterValue() string { return p.Name }

type taskItem struct {
	models.Task
}

func (t taskItem) Title() string {
	var statusSym string
	switch t.Task.Status {
	case "todo":
		statusSym = "[]"
	case "in_progress":
		statusSym = "[in_progress]"
	case "done":
		statusSym = "[x]"
	default:
		statusSym = "[" + t.Task.Status + "]"
	}
	s := statusSym
	if t.Task.Priority != "" {
		s += " " + t.Task.Priority + " "
	}
	return s + " " + t.Task.Title
}
func (t taskItem) Description() string {
	if t.Task.DueDate != nil {
		return "due: " + t.Task.DueDate.Format("2006-01-02")
	}
	return t.Task.Description
}
func (t taskItem) FilterValue() string { return t.Task.Title }

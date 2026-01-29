package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/cli-todo/internal/store"
)

// refreshMsg forces the list to refresh on the next update.
type refreshMsg struct{}

func (m *model) refreshList() tea.Cmd {
	switch m.screen {
	case screenWorkspaces:
		ws, err := store.ListWorkspaces(m.db)
		if err != nil {
			m.err = err.Error()
			return nil
		}
		m.workspaces = ws
		m.err = ""
		if m.workspaceCursor >= len(m.workspaces) {
			m.workspaceCursor = max(0, len(m.workspaces)-1)
		}
		return nil
	case screenProjects:
		if m.selectedWorkspace == nil {
			return nil
		}
		projs, err := store.ListProjects(m.db, m.selectedWorkspace.ID)
		if err != nil {
			m.err = err.Error()
			return nil
		}
		m.projects = projs
		m.err = ""
		items := make([]list.Item, 0, len(projs)+1)
		items = append(items, projectItem{Name: "Default", IsDefault: true})
		for i := range projs {
			items = append(items, projectItem{ID: &projs[i].ID, Name: projs[i].Name, Color: projs[i].Color})
		}
		m.setBubblesList(" "+m.selectedWorkspace.Name+" → Lists ", items)
		return nil
	case screenTasks:
		if m.selectedWorkspace == nil {
			return nil
		}
		tasks, err := store.ListTasks(m.db, m.selectedWorkspace.ID, m.selectedProjectID)
		if err != nil {
			m.err = err.Error()
			return nil
		}
		m.tasks = tasks
		m.err = ""
		items := make([]list.Item, len(tasks))
		for i := range tasks {
			items[i] = taskItem{Task: tasks[i]}
		}
		title := " Default "
		if m.selectedProjectID != nil {
			for _, p := range m.projects {
				if p.ID == *m.selectedProjectID {
					title = " " + p.Name + " "
					break
				}
			}
		}
		m.setBubblesList(m.selectedWorkspace.Name+" →"+title+" Tasks ", items)
		return nil
	}
	return nil
}

// setBubblesList creates or updates the list component with the given title and items.
func (m *model) setBubblesList(title string, items []list.Item) {
	if m.list.Title == title {
		m.list.SetItems(items)
		return
	}
	delegate := list.NewDefaultDelegate()
	m.list = list.New(items, delegate, m.width, max(1, m.height-4))
	m.list.Title = title
	m.list.SetShowHelp(false)
	m.list.SetFilteringEnabled(true)
	m.list.DisableQuitKeybindings()
}

func (m *model) handleBack() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenProjects:
		if m.moveTaskID != 0 {
			m.moveTaskID = 0
			m.screen = screenTasks
		} else {
			m.screen = screenWorkspaces
			m.selectedWorkspace = nil
			m.selectedProjectID = nil
		}
	case screenTasks:
		m.screen = screenProjects
		m.selectedProjectID = nil
	default:
		return m, nil
	}
	return m, m.refreshList()
}

func (m *model) handleAdd() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenWorkspaces:
		m.inputMode = inputNewWorkspace
	case screenProjects:
		m.inputMode = inputNewProject
	case screenTasks:
		m.inputMode = inputNewTask
		m.clearNewTaskDraft()
		m.input.Placeholder = "Title (Enter = quick add, Tab = more options)"
	default:
		return m, nil
	}
	m.input.SetValue("")
	m.input.Focus()
	return m, textinput.Blink
}

func (m *model) clearNewTaskDraft() {
	m.newTaskTitle = ""
	m.newTaskDue = ""
	m.newTaskPriority = ""
	m.newTaskStatus = ""
}

func normalizePriority(s string) string {
	switch s {
	case "low", "medium", "high":
		return s
	case "l":
		return "low"
	case "m":
		return "medium"
	case "h":
		return "high"
	}
	return ""
}

func normalizeStatus(s string) string {
	switch s {
	case "todo", "in_progress", "done":
		return s
	case "t":
		return "todo"
	case "i":
		return "in_progress"
	case "d":
		return "done"
	}
	return "todo"
}

// createTaskFromDraft creates a task from the new-task draft (title + optional due, priority, status).
func (m *model) createTaskFromDraft(statusInput string) (tea.Model, tea.Cmd) {
	status := strings.TrimSpace(statusInput)
	if status == "" {
		status = "todo"
	}
	status = normalizeStatus(status)
	priority := normalizePriority(m.newTaskPriority)
	var due *time.Time
	if m.newTaskDue != "" {
		for _, layout := range []string{"2006-01-02", "2006/01/02"} {
			if t, err := time.Parse(layout, m.newTaskDue); err == nil {
				due = &t
				break
			}
		}
		if due == nil {
			m.err = "Invalid date (use YYYY-MM-DD)"
			return m, nil
		}
	}
	if _, err := store.CreateTask(m.db, m.selectedWorkspace.ID, m.selectedProjectID, m.newTaskTitle, "", status, priority, due); err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.err = ""
	m.clearNewTaskDraft()
	m.input.SetValue("")
	m.inputMode = inputNone
	m.input.Placeholder = "Name..."
	return m, m.refreshList()
}

func (m *model) handleEdit() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenWorkspaces:
		if m.workspaceCursor < 0 || m.workspaceCursor >= len(m.workspaces) {
			return m, nil
		}
		w := m.workspaces[m.workspaceCursor]
		m.editWorkspaceID = w.ID
		m.inputMode = inputEditWorkspace
		m.input.SetValue(w.Name)
		m.input.Focus()
		return m, textinput.Blink
	case screenProjects:
		sel := m.list.SelectedItem()
		if sel == nil {
			return m, nil
		}
		p, ok := sel.(projectItem)
		if !ok || p.IsDefault || p.ID == nil {
			return m, nil
		}
		m.editProjectID = *p.ID
		m.inputMode = inputEditProject
		m.input.SetValue(p.Name)
		m.input.Focus()
		return m, textinput.Blink
	case screenTasks:
		sel := m.list.SelectedItem()
		if sel == nil {
			return m, nil
		}
		t, ok := sel.(taskItem)
		if !ok {
			return m, nil
		}
		m.editTaskID = t.ID
		m.inputMode = inputEditTask
		m.input.SetValue(t.Task.Title)
		m.input.Focus()
		return m, textinput.Blink
	}
	return m, nil
}

func (m *model) handleWorkspaceColor() (tea.Model, tea.Cmd) {
	if m.workspaceCursor < 0 || m.workspaceCursor >= len(m.workspaces) {
		return m, nil
	}
	w := m.workspaces[m.workspaceCursor]
	m.editWorkspaceID = w.ID
	m.inputMode = inputWorkspaceColor
	m.input.SetValue(w.Color)
	m.input.Focus()
	return m, textinput.Blink
}

func (m *model) handleProjectColor() (tea.Model, tea.Cmd) {
	sel := m.list.SelectedItem()
	if sel == nil {
		return m, nil
	}
	p, ok := sel.(projectItem)
	if !ok || p.IsDefault || p.ID == nil {
		return m, nil
	}
	m.editProjectID = *p.ID
	m.inputMode = inputProjectColor
	m.input.SetValue(p.Color)
	m.input.Focus()
	return m, textinput.Blink
}

func (m *model) handleMoveTask() (tea.Model, tea.Cmd) {
	t, ok := m.getSelectedTask()
	if !ok {
		return m, nil
	}
	m.moveTaskID = t.ID
	m.screen = screenProjects
	return m, m.refreshList()
}

func (m *model) getSelectedTask() (taskItem, bool) {
	sel := m.list.SelectedItem()
	if sel == nil {
		return taskItem{}, false
	}
	t, ok := sel.(taskItem)
	return t, ok
}

var statusOrder = []string{"todo", "in_progress", "done"}

func (m *model) handleTaskCycleStatus() (tea.Model, tea.Cmd) {
	t, ok := m.getSelectedTask()
	if !ok {
		return m, nil
	}
	next := t.Task.Status
	for i, s := range statusOrder {
		if s == t.Task.Status {
			next = statusOrder[(i+1)%len(statusOrder)]
			break
		}
	}
	task, err := store.GetTask(m.db, t.ID)
	if err != nil {
		m.err = err.Error()
		return m, nil
	}
	if _, err := store.UpdateTask(m.db, t.ID, task.Title, task.Description, next, task.Priority, task.DueDate); err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.err = ""
	m.statusMsg = "Status: " + next
	return m, m.refreshList()
}

var priorityOrder = []string{"", "low", "medium", "high"}

func (m *model) handleTaskCyclePriority() (tea.Model, tea.Cmd) {
	t, ok := m.getSelectedTask()
	if !ok {
		return m, nil
	}
	task, err := store.GetTask(m.db, t.ID)
	if err != nil {
		m.err = err.Error()
		return m, nil
	}
	cur := task.Priority
	next := ""
	for i, p := range priorityOrder {
		if p == cur {
			next = priorityOrder[(i+1)%len(priorityOrder)]
			break
		}
	}
	if _, err := store.UpdateTask(m.db, t.ID, task.Title, task.Description, task.Status, next, task.DueDate); err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.err = ""
	if next == "" {
		m.statusMsg = "Priority cleared"
	} else {
		m.statusMsg = "Priority: " + next
	}
	return m, m.refreshList()
}

func (m *model) handleTaskSetDueDate() (tea.Model, tea.Cmd) {
	t, ok := m.getSelectedTask()
	if !ok {
		return m, nil
	}
	m.editTaskID = t.ID
	m.inputMode = inputTaskDueDate
	if t.Task.DueDate != nil {
		m.input.SetValue(t.Task.DueDate.Format("2006-01-02"))
	} else {
		m.input.SetValue("")
	}
	m.input.Focus()
	return m, textinput.Blink
}

func (m *model) handleInputSubmit(val string, mode inputKind) (*model, tea.Cmd) {
	switch mode {
	case inputNewWorkspace:
		if _, err := store.CreateWorkspace(m.db, val); err != nil {
			m.err = err.Error()
			m.statusMsg = ""
			return m, nil
		}
		m.err = ""
		m.statusMsg = "Added: " + val
		m.refreshList()
		return m, func() tea.Msg { return refreshMsg{} }
	case inputNewProject:
		if m.selectedWorkspace == nil {
			return m, nil
		}
		if _, err := store.CreateProject(m.db, m.selectedWorkspace.ID, val); err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.err = ""
		return m, m.refreshList()
	case inputNewTask:
		if m.selectedWorkspace == nil {
			return m, nil
		}
		if _, err := store.CreateTask(m.db, m.selectedWorkspace.ID, m.selectedProjectID, val, "", "todo", "", nil); err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.err = ""
		return m, m.refreshList()
	case inputEditWorkspace:
		if m.editWorkspaceID == 0 {
			return m, nil
		}
		if _, err := store.UpdateWorkspace(m.db, m.editWorkspaceID, val); err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.err = ""
		m.editWorkspaceID = 0
		m.statusMsg = "Updated"
		m.refreshList()
		return m, func() tea.Msg { return refreshMsg{} }
	case inputEditProject:
		if m.editProjectID == 0 {
			return m, nil
		}
		if _, err := store.UpdateProject(m.db, m.editProjectID, val); err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.err = ""
		m.editProjectID = 0
		m.statusMsg = "Updated"
		return m, m.refreshList()
	case inputEditTask:
		if m.editTaskID == 0 {
			return m, nil
		}
		task, err := store.GetTask(m.db, m.editTaskID)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		if _, err := store.UpdateTask(m.db, m.editTaskID, val, task.Description, task.Status, task.Priority, task.DueDate); err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.err = ""
		m.editTaskID = 0
		m.statusMsg = "Updated"
		return m, m.refreshList()
	case inputTaskDueDate:
		if m.editTaskID == 0 {
			return m, nil
		}
		task, err := store.GetTask(m.db, m.editTaskID)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		var due *time.Time
		if val != "" {
			for _, layout := range []string{"2006-01-02", "2006/01/02"} {
				if t, err := time.Parse(layout, val); err == nil {
					due = &t
					break
				}
			}
			if due == nil {
				m.err = "Invalid date (use YYYY-MM-DD)"
				m.editTaskID = 0
				return m, nil
			}
		}
		if _, err := store.UpdateTask(m.db, m.editTaskID, task.Title, task.Description, task.Status, task.Priority, due); err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.err = ""
		m.editTaskID = 0
		if due == nil {
			m.statusMsg = "Due date cleared"
		} else {
			m.statusMsg = "Due: " + due.Format("2006-01-02")
		}
		return m, m.refreshList()
	case inputWorkspaceColor:
		if m.editWorkspaceID == 0 {
			return m, nil
		}
		if _, err := store.SetWorkspaceColor(m.db, m.editWorkspaceID, val); err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.err = ""
		m.editWorkspaceID = 0
		m.statusMsg = "Color set"
		m.refreshList()
		return m, func() tea.Msg { return refreshMsg{} }
	case inputProjectColor:
		if m.editProjectID == 0 {
			return m, nil
		}
		if _, err := store.SetProjectColor(m.db, m.editProjectID, val); err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.err = ""
		m.editProjectID = 0
		m.statusMsg = "Color set"
		return m, m.refreshList()
	}
	return m, nil
}

func (m *model) handleDelete() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenWorkspaces:
		if m.workspaceCursor < 0 || m.workspaceCursor >= len(m.workspaces) {
			return m, nil
		}
		w := m.workspaces[m.workspaceCursor]
		_ = store.DeleteWorkspace(m.db, w.ID)
		if m.workspaceCursor >= len(m.workspaces)-1 {
			m.workspaceCursor = max(0, len(m.workspaces)-2)
		}
		return m, m.refreshList()
	case screenProjects:
		sel := m.list.SelectedItem()
		if sel == nil {
			return m, nil
		}
		p, ok := sel.(projectItem)
		if !ok || p.IsDefault || p.ID == nil {
			return m, nil
		}
		_ = store.DeleteProject(m.db, *p.ID)
		return m, m.refreshList()
	case screenTasks:
		sel := m.list.SelectedItem()
		if sel == nil {
			return m, nil
		}
		t, ok := sel.(taskItem)
		if !ok {
			return m, nil
		}
		_ = store.DeleteTask(m.db, t.ID)
		return m, m.refreshList()
	}
	return m, nil
}

func (m *model) handleSelect() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenWorkspaces:
		if m.workspaceCursor < 0 || m.workspaceCursor >= len(m.workspaces) {
			return m, nil
		}
		m.selectedWorkspace = &m.workspaces[m.workspaceCursor]
		m.screen = screenProjects
		return m, m.refreshList()
	case screenProjects:
		sel := m.list.SelectedItem()
		if sel == nil {
			return m, nil
		}
		p, ok := sel.(projectItem)
		if !ok {
			return m, nil
		}
		if m.moveTaskID != 0 {
			_, err := store.SetTaskProject(m.db, m.moveTaskID, p.ID)
			if err != nil {
				m.err = err.Error()
				return m, nil
			}
			m.err = ""
			m.statusMsg = "Task moved"
			m.moveTaskID = 0
			m.selectedProjectID = p.ID
			m.screen = screenTasks
			return m, m.refreshList()
		}
		m.selectedProjectID = p.ID
		m.screen = screenTasks
		return m, m.refreshList()
	case screenTasks:
		// e = edit task (cycle status) could go here
		return m, nil
	}
	return m, nil
}

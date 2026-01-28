package tui

import (
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
			items = append(items, projectItem{ID: &projs[i].ID, Name: projs[i].Name})
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
		m.screen = screenWorkspaces
		m.selectedWorkspace = nil
		m.selectedProjectID = nil
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
	default:
		return m, nil
	}
	m.input.SetValue("")
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
		m.selectedProjectID = p.ID
		m.screen = screenTasks
		return m, m.refreshList()
	case screenTasks:
		// e = edit task (cycle status) could go here
		return m, nil
	}
	return m, nil
}

package tui

import (
	"database/sql"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli-todo/internal/models"
	"github.com/cli-todo/internal/store"
)

type screen int

const (
	screenWorkspaces screen = iota
	screenProjects
	screenTasks
)

type inputKind int

const (
	inputNone inputKind = iota
	inputNewWorkspace
	inputNewProject
	inputNewTask
)

type model struct {
	db        *sql.DB
	screen    screen
	width     int
	height    int
	list      list.Model
	workspaces []models.Workspace
	projects   []models.Project
	tasks     []models.Task
	selectedWorkspace *models.Workspace
	selectedProjectID  *int64 // nil = default list
	inputMode inputKind
	input     textinput.Model
	err       string
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
)

func New(db *sql.DB) *model {
	ti := textinput.New()
	ti.Placeholder = "Name..."
	ti.CharLimit = 200
	ti.Width = 40
	return &model{db: db, screen: screenWorkspaces, input: ti}
}

func (m *model) Init() tea.Cmd {
	m.width = 60
	m.height = 20
	return m.refreshList()
}

const workspacesListTitle = " Workspaces "

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
		items := make([]list.Item, len(ws))
		for i := range ws {
			items[i] = workspaceItem{Workspace: ws[i]}
		}
		if m.list.Title == workspacesListTitle {
			m.list.SetItems(items)
		} else {
			delegate := list.NewDefaultDelegate()
			m.list = list.New(items, delegate, m.width, max(1, m.height-4))
			m.list.Title = workspacesListTitle
			m.list.SetShowHelp(false)
			m.list.SetFilteringEnabled(true)
			m.list.DisableQuitKeybindings()
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
		projTitle := " " + m.selectedWorkspace.Name + " → Lists "
		if m.list.Title == projTitle {
			m.list.SetItems(items)
		} else {
			delegate := list.NewDefaultDelegate()
			m.list = list.New(items, delegate, m.width, max(1, m.height-4))
			m.list.Title = projTitle
			m.list.SetShowHelp(false)
			m.list.SetFilteringEnabled(true)
			m.list.DisableQuitKeybindings()
		}
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
		delegate := list.NewDefaultDelegate()
		m.list = list.New(items, delegate, m.width, max(1, m.height-4))
		title := " Default "
		if m.selectedProjectID != nil {
			for _, p := range m.projects {
				if p.ID == *m.selectedProjectID {
					title = " " + p.Name + " "
					break
				}
			}
		}
		taskTitle := m.selectedWorkspace.Name + " →" + title + " Tasks "
		if m.list.Title == taskTitle {
			m.list.SetItems(items)
		} else {
			m.list = list.New(items, delegate, m.width, max(1, m.height-4))
			m.list.Title = taskTitle
			m.list.SetShowHelp(false)
			m.list.SetFilteringEnabled(true)
			m.list.DisableQuitKeybindings()
		}
		return nil
	}
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.inputMode != inputNone {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				val := strings.TrimSpace(m.input.Value())
				m.input.SetValue("")
				m.inputMode = inputNone
				if val == "" {
					return m, nil
				}
				var cmd tea.Cmd
				var next *model
				next, cmd = m.handleInputSubmit(val)
				return next, tea.Batch(cmd)
			case "esc", "ctrl+c":
				m.inputMode = inputNone
				m.input.SetValue("")
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "backspace":
			return m.handleBack()
		case "a":
			return m.handleAdd()
		case "d":
			return m.handleDelete()
		case "enter":
			return m.handleSelect()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
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
		m.input.SetValue("")
		m.input.Focus()
		return m, textinput.Blink
	case screenProjects:
		m.inputMode = inputNewProject
		m.input.SetValue("")
		m.input.Focus()
		return m, textinput.Blink
	case screenTasks:
		m.inputMode = inputNewTask
		m.input.SetValue("")
		m.input.Focus()
		return m, textinput.Blink
	}
	return m, nil
}

func (m *model) handleInputSubmit(val string) (*model, tea.Cmd) {
	switch m.inputMode {
	case inputNewWorkspace:
		_, err := store.CreateWorkspace(m.db, val)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.err = ""
		return m, m.refreshList()
	case inputNewProject:
		if m.selectedWorkspace == nil {
			return m, nil
		}
		_, err := store.CreateProject(m.db, m.selectedWorkspace.ID, val)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.err = ""
		return m, m.refreshList()
	case inputNewTask:
		if m.selectedWorkspace == nil {
			return m, nil
		}
		_, err := store.CreateTask(m.db, m.selectedWorkspace.ID, m.selectedProjectID, val, "", "todo", "", nil)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.err = ""
		return m, m.refreshList()
	}
	return m, nil
}

func (m *model) handleDelete() (tea.Model, tea.Cmd) {
	sel := m.list.SelectedItem()
	if sel == nil {
		return m, nil
	}
	switch m.screen {
	case screenWorkspaces:
		if w, ok := sel.(workspaceItem); ok {
			_ = store.DeleteWorkspace(m.db, w.ID)
			return m, m.refreshList()
		}
	case screenProjects:
		if p, ok := sel.(projectItem); ok && p.IsDefault {
			return m, nil
		}
		if p, ok := sel.(projectItem); ok && p.ID != nil {
			_ = store.DeleteProject(m.db, *p.ID)
			return m, m.refreshList()
		}
	case screenTasks:
		if t, ok := sel.(taskItem); ok {
			_ = store.DeleteTask(m.db, t.ID)
			return m, m.refreshList()
		}
	}
	return m, nil
}

func (m *model) handleSelect() (tea.Model, tea.Cmd) {
	sel := m.list.SelectedItem()
	if sel == nil {
		return m, nil
	}
	switch m.screen {
	case screenWorkspaces:
		if w, ok := sel.(workspaceItem); ok {
			m.selectedWorkspace = &w.Workspace
			m.screen = screenProjects
			return m, m.refreshList()
		}
	case screenProjects:
		if p, ok := sel.(projectItem); ok {
			m.selectedProjectID = p.ID
			m.screen = screenTasks
			return m, m.refreshList()
		}
	case screenTasks:
		// optional: e = edit task (cycle status); for now no-op
		return m, nil
	}
	return m, nil
}

func (m *model) View() string {
	if m.inputMode != inputNone {
		prompt := "Name: "
		switch m.inputMode {
		case inputNewWorkspace:
			prompt = "New workspace name: "
		case inputNewProject:
			prompt = "New project/list name: "
		case inputNewTask:
			prompt = "New task title: "
		}
		return titleStyle.Render("Todo") + "\n\n" + prompt + m.input.View() + "\n\n" + helpStyle.Render("Enter confirm, Esc cancel")
	}
	s := titleStyle.Render("Todo") + "\n\n"
	s += m.list.View()
	if m.err != "" {
		s += "\n" + errorStyle.Render("Error: "+m.err)
	}
	s += "\n" + helpStyle.Render("↑/↓ move • Enter open/select • a add • d delete • ← back • q quit")
	return s
}

// Run starts the TUI program.
func Run(db *sql.DB) error {
	p := tea.NewProgram(New(db), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

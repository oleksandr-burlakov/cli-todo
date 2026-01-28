package tui

import (
	"database/sql"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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
	db                *sql.DB
	dbPath            string
	screen            screen
	width             int
	height            int
	list              list.Model
	workspaces        []models.Workspace
	workspaceCursor   int
	projects          []models.Project
	tasks             []models.Task
	selectedWorkspace *models.Workspace
	selectedProjectID *int64
	inputMode         inputKind
	input             textinput.Model
	err               string
	statusMsg         string
}

func New(db *sql.DB) *model {
	path, err := store.DBPath()
	if err != nil {
		path = "?"
	}
	ti := textinput.New()
	ti.Placeholder = "Name..."
	ti.CharLimit = 200
	ti.Width = 40
	return &model{db: db, dbPath: path, screen: screenWorkspaces, input: ti}
}

func (m *model) Init() tea.Cmd {
	m.width = 60
	m.height = 20
	return m.refreshList()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.inputMode != inputNone {
		return m.updateInput(msg)
	}
	switch msg := msg.(type) {
	case refreshMsg:
		m.statusMsg = ""
		return m, m.refreshList()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.screen != screenWorkspaces {
			m.list.SetSize(msg.Width, msg.Height-4)
		}
		return m, nil
	case tea.KeyMsg:
		k := msg.String()
		if k == "ctrl+c" || k == "q" {
			return m, tea.Quit
		}
		if k == "backspace" || k == "ctrl+h" || k == "left" {
			return m.handleBack()
		}
		if k == "a" {
			return m.handleAdd()
		}
		if k == "d" {
			return m.handleDelete()
		}
		if k == "enter" {
			return m.handleSelect()
		}
		if m.screen == screenWorkspaces {
			return m.updateWorkspaceNav(k)
		}
	}
	var cmd tea.Cmd
	if m.screen != screenWorkspaces {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m *model) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		if k == "enter" || k == "ctrl+j" || k == "ctrl+m" {
			val := strings.TrimSpace(m.input.Value())
			m.input.SetValue("")
			mode := m.inputMode // save before clearing
			m.inputMode = inputNone
			if val == "" {
				return m, nil
			}
			next, cmd := m.handleInputSubmit(val, mode)
			return next, cmd
		}
		if k == "esc" || k == "ctrl+c" {
			m.inputMode = inputNone
			m.input.SetValue("")
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *model) updateWorkspaceNav(k string) (tea.Model, tea.Cmd) {
	if k == "up" || k == "k" {
		if m.workspaceCursor > 0 {
			m.workspaceCursor--
		}
		return m, nil
	}
	if k == "down" || k == "j" {
		if m.workspaceCursor < len(m.workspaces)-1 {
			m.workspaceCursor++
		}
		return m, nil
	}
	return m, nil
}

func (m *model) View() string {
	if m.inputMode != inputNone {
		return m.viewInput()
	}
	s := m.viewContent()
	if m.err != "" {
		s += "\n" + errorStyle.Render("Error: "+m.err)
	}
	s += "\n" + m.viewFooter()
	return s
}

// Run starts the TUI program.
func Run(db *sql.DB) error {
	p := tea.NewProgram(New(db), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

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
	"github.com/muesli/termenv"
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
	inputNewTaskDue
	inputNewTaskPriority
	inputNewTaskStatus
	inputEditWorkspace
	inputEditProject
	inputEditTask
	inputTaskDueDate
	inputWorkspaceColor
	inputProjectColor
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
	editWorkspaceID   int64 // 0 = not editing
	editProjectID     int64
	editTaskID        int64
	moveTaskID        int64 // when non-zero, selecting project to move task to
	// Draft for multi-step new task (title required; due, priority, status optional)
	newTaskTitle    string
	newTaskDue      string
	newTaskPriority string
	newTaskStatus   string
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
		if k == "e" {
			return m.handleEdit()
		}
		if m.screen == screenTasks {
			if k == "s" {
				return m.handleTaskCycleStatus()
			}
			if k == "p" {
				return m.handleTaskCyclePriority()
			}
			if k == "u" {
				return m.handleTaskSetDueDate()
			}
			if k == "m" {
				return m.handleMoveTask()
			}
		}
		if m.screen == screenWorkspaces {
			if k == "c" {
				return m.handleWorkspaceColor()
			}
			return m.updateWorkspaceNav(k)
		}
		if m.screen == screenProjects {
			if k == "c" {
				return m.handleProjectColor()
			}
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
			mode := m.inputMode

			// Multi-step new task: quick create (Enter on title) or advance / create from form
			switch mode {
			case inputNewTask:
				if val == "" {
					return m, nil
				}
				m.input.SetValue("")
				m.inputMode = inputNone
				next, cmd := m.handleInputSubmit(val, inputNewTask)
				return next, cmd
			case inputNewTaskDue:
				m.newTaskDue = val
				m.input.SetValue("")
				m.input.Placeholder = "Priority: low / medium / high (optional)"
				m.inputMode = inputNewTaskPriority
				return m, textinput.Blink
			case inputNewTaskPriority:
				m.newTaskPriority = val
				m.input.SetValue("")
				m.input.Placeholder = "Status: todo / in_progress / done (optional)"
				m.inputMode = inputNewTaskStatus
				return m, textinput.Blink
			case inputNewTaskStatus:
				return m.createTaskFromDraft(val)
			}

			// Single-step submit for all other modes
			if val == "" && mode != inputTaskDueDate && mode != inputWorkspaceColor && mode != inputProjectColor {
				return m, nil
			}
			m.input.SetValue("")
			m.inputMode = inputNone
			next, cmd := m.handleInputSubmit(val, mode)
			return next, cmd
		}
		if k == "tab" {
			mode := m.inputMode
			switch mode {
			case inputNewTask:
				val := strings.TrimSpace(m.input.Value())
				if val == "" {
					return m, nil
				}
				m.newTaskTitle = val
				m.input.SetValue("")
				m.input.Placeholder = "Due date (YYYY-MM-DD, optional)"
				m.inputMode = inputNewTaskDue
				return m, textinput.Blink
			case inputNewTaskDue:
				m.newTaskDue = strings.TrimSpace(m.input.Value())
				m.input.SetValue("")
				m.input.Placeholder = "Priority: low / medium / high (optional)"
				m.inputMode = inputNewTaskPriority
				return m, textinput.Blink
			case inputNewTaskPriority:
				m.newTaskPriority = strings.TrimSpace(m.input.Value())
				m.input.SetValue("")
				m.input.Placeholder = "Status: todo / in_progress / done (optional)"
				m.inputMode = inputNewTaskStatus
				return m, textinput.Blink
			case inputNewTaskStatus:
				return m.createTaskFromDraft(strings.TrimSpace(m.input.Value()))
			}
		}
		if k == "esc" || k == "ctrl+c" {
			m.inputMode = inputNone
			m.input.SetValue("")
			m.clearNewTaskDraft()
			m.input.Placeholder = "Name..."
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
	// Force color output so workspace/project colors render in the terminal.
	lipgloss.SetColorProfile(termenv.TrueColor)
	p := tea.NewProgram(New(db), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

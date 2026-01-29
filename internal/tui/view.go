package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// namedColors maps common names to hex for lipgloss (which expects hex with # or ANSI numbers).
var namedColors = map[string]string{
	"green": "#00ff00", "red": "#ff0000", "blue": "#0000ff", "yellow": "#ffff00",
	"cyan": "#00ffff", "magenta": "#ff00ff", "white": "#ffffff", "black": "#000000",
	"orange": "#ffa500", "purple": "#800080", "pink": "#ff69b4", "gray": "#808080", "grey": "#808080",
}

// lipglossColor normalizes a color string for lipgloss: hex with #, or named -> hex.
func lipglossColor(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return ""
	}
	if hex, ok := namedColors[s]; ok {
		return hex
	}
	if s[0] == '#' {
		return s
	}
	// Allow 6-char hex without # (e.g. ff0000)
	if len(s) == 6 && isHex(s) {
		return "#" + s
	}
	return s
}

func isHex(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func (m *model) viewInput() string {
	prompt := "Name: "
	switch m.inputMode {
	case inputNewWorkspace:
		prompt = "New workspace name: "
	case inputNewProject:
		prompt = "New project/list name: "
	case inputNewTask:
		prompt = "New task title: "
	case inputNewTaskDue:
		prompt = "Due date (YYYY-MM-DD, optional): "
	case inputNewTaskPriority:
		prompt = "Priority (low/medium/high, optional): "
	case inputNewTaskStatus:
		prompt = "Status (todo/in_progress/done, optional): "
	case inputEditWorkspace:
		prompt = "Edit workspace name: "
	case inputEditProject:
		prompt = "Edit project/list name: "
	case inputEditTask:
		prompt = "Edit task title: "
	case inputTaskDueDate:
		prompt = "Due date (YYYY-MM-DD, or leave empty to clear): "
	case inputWorkspaceColor:
		prompt = "Workspace color (e.g. green, blue, #ff0000; empty to clear): "
	case inputProjectColor:
		prompt = "Project color (e.g. green, blue, #ff0000; empty to clear): "
	}
	help := "Press Enter to save • Esc to cancel"
	if m.inputMode == inputNewTask || m.inputMode == inputNewTaskDue || m.inputMode == inputNewTaskPriority || m.inputMode == inputNewTaskStatus {
		help = "Enter = create or next • Tab = next field • Esc = cancel"
	}
	return titleStyle.Render("Todo") + "\n\n" + prompt + m.input.View() + "\n\n" + helpStyle.Render(help)
}

func (m *model) viewContent() string {
	s := titleStyle.Render("Todo") + "\n\n"
	if m.screen == screenWorkspaces {
		s += titleStyle.Render(" Workspaces ") + "\n\n"
		if len(m.workspaces) == 0 {
			s += helpStyle.Render("  (none yet — press 'a' to add one)") + "\n"
		} else {
			for i, w := range m.workspaces {
				cursor := "  "
				if i == m.workspaceCursor {
					cursor = "> "
				}
				name := w.Name
				if w.Color != "" {
					c := lipglossColor(w.Color)
					if c != "" {
						name = lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(w.Name)
					}
				}
				s += cursor + name + "\n"
			}
		}
	} else if m.screen == screenProjects {
		title := m.selectedWorkspace.Name + " → Lists "
		if m.moveTaskID != 0 {
			title = "Move task to project "
		}
		s += titleStyle.Render(title) + "\n\n"
		items := m.list.VisibleItems()
		idx := m.list.Index()
		for i, item := range items {
			p, ok := item.(projectItem)
			if !ok {
				continue
			}
			cursor := "  "
			if i == idx {
				cursor = "> "
			}
			name := p.Name
			if p.Color != "" {
				c := lipglossColor(p.Color)
				if c != "" {
					name = lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(p.Name)
				}
			}
			s += cursor + name + "\n"
		}
	} else {
		s += m.list.View()
	}
	return s
}

func (m *model) viewFooter() string {
	help := "↑/↓ move • Enter open/select • a add • e edit • c color • d delete • ← back • q quit"
	if m.screen == screenProjects {
		if m.moveTaskID != 0 {
			help = "↑/↓ move • Enter move here • ← cancel"
		} else {
			help = "↑/↓ move • Enter open/select • a add • e edit • c color • d delete • ← back • q quit"
		}
	}
	if m.screen == screenTasks {
		help = "↑/↓ move • Enter select • a add • e edit • s status • p priority • u due date • m move • d delete • ← back • q quit"
	}
	s := helpStyle.Render(help)
	if m.statusMsg != "" {
		s += "\n" + statusStyle.Render(m.statusMsg)
	}
	dbLine := "DB: " + m.dbPath
	if m.screen == screenWorkspaces {
		dbLine += "  •  Workspaces: " + fmt.Sprintf("%d", len(m.workspaces))
	}
	s += "\n" + helpStyle.Render(dbLine)
	return s
}

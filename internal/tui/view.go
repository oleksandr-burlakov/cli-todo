package tui

import "fmt"

func (m *model) viewInput() string {
	prompt := "Name: "
	switch m.inputMode {
	case inputNewWorkspace:
		prompt = "New workspace name: "
	case inputNewProject:
		prompt = "New project/list name: "
	case inputNewTask:
		prompt = "New task title: "
	}
	return titleStyle.Render("Todo") + "\n\n" + prompt + m.input.View() + "\n\n" + helpStyle.Render("Press Enter to save • Esc to cancel")
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
				s += cursor + w.Name + "\n"
			}
		}
	} else {
		s += m.list.View()
	}
	return s
}

func (m *model) viewFooter() string {
	s := helpStyle.Render("↑/↓ move • Enter open/select • a add • d delete • ← back • q quit")
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

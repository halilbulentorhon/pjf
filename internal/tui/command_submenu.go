package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/halilbulentorhon/pjf/internal/project"
	"github.com/halilbulentorhon/pjf/internal/service"
)

type cmdSubmenuResult struct {
	command string
	action  string
}

type cmdSubmenuModel struct {
	project  project.Project
	commands []service.ResolvedCommand
	cursor   int
}

func newCmdSubmenuModel(p project.Project, svc *service.ProjectService) cmdSubmenuModel {
	return cmdSubmenuModel{
		project:  p,
		commands: svc.ResolveCommands(p),
	}
}

func (m cmdSubmenuModel) totalItems() int {
	return len(m.commands) + 1
}

func (m cmdSubmenuModel) Update(msg tea.Msg) (cmdSubmenuModel, tea.Cmd, cmdSubmenuResult) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < m.totalItems()-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.commands) {
				cmd := m.commands[m.cursor]
				return m, nil, cmdSubmenuResult{command: cmd.Command, action: "run"}
			}
			if m.cursor == len(m.commands) {
				return m, nil, cmdSubmenuResult{action: "custom-input"}
			}
		}
	}
	return m, nil, cmdSubmenuResult{}
}

func (m cmdSubmenuModel) View() string {
	return actionMenuStyle.Render(func() string {
		s := titleStyle.Render("Run Command") + "\n\n"

		for i, cmd := range m.commands {
			label := cmd.Name
			if cmd.IsProject {
				label += " " + helpStyle.Render("(project)")
			} else {
				label += " " + helpStyle.Render("(global)")
			}
			if i == m.cursor {
				s += selectedItemStyle.Render("▸ "+label) + "\n"
			} else {
				s += itemStyle.Render("  "+label) + "\n"
			}
		}

		if len(m.commands) > 0 {
			s += separatorStyle.Render("  ──────────────") + "\n"
		}

		customLabel := "Run custom command..."
		if m.cursor == len(m.commands) {
			s += selectedItemStyle.Render("▸ "+customLabel) + "\n"
		} else {
			s += itemStyle.Render("  "+customLabel) + "\n"
		}

		s += "\n" + helpStyle.Render("enter: run  esc: back")
		return s
	}())
}

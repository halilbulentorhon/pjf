package tui

import (
	"fmt"

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
		key := msg.String()
		if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
			idx := int(key[0] - '1')
			if idx < m.totalItems() {
				m.cursor = idx
				key = "enter"
			}
		}
		switch key {
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
			num := fmt.Sprintf("%d. ", i+1)
			label := cmd.Name
			if cmd.IsProject {
				label += " " + helpStyle.Render("(project)")
			} else {
				label += " " + helpStyle.Render("(global)")
			}
			if i == m.cursor {
				s += selectedItemStyle.Render("▸ "+num+label) + "\n"
			} else {
				s += itemStyle.Render("  "+num+label) + "\n"
			}
		}

		if len(m.commands) > 0 {
			s += separatorStyle.Render("  ──────────────") + "\n"
		}

		customNum := fmt.Sprintf("%d. ", len(m.commands)+1)
		customLabel := "Run custom command..."
		if m.cursor == len(m.commands) {
			s += selectedItemStyle.Render("▸ "+customNum+customLabel) + "\n"
		} else {
			s += itemStyle.Render("  "+customNum+customLabel) + "\n"
		}

		s += "\n" + helpStyle.Render("1-9: select  enter: run  esc: back")
		return s
	}())
}

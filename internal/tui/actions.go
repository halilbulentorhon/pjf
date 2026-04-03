package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/halilbulentorhon/pjf/internal/project"
	"github.com/halilbulentorhon/pjf/internal/service"
)

type actionItem struct {
	label string
	run   func() error
}

type actionsModel struct {
	project project.Project
	items   []actionItem
	cursor  int
}

func newActionsModel() actionsModel {
	return actionsModel{}
}

func newActionsModelForProject(p project.Project, svc *service.ProjectService) actionsModel {
	items := []actionItem{
		{label: "Open in Terminal", run: func() error { return svc.OpenTerminal(p) }},
		{label: "Copy Path", run: func() error { return svc.CopyPath(p) }},
	}
	return actionsModel{
		project: p,
		items:   items,
	}
}

func (m actionsModel) Update(msg tea.Msg) (actionsModel, tea.Cmd, string) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.items) {
				item := m.items[m.cursor]
				if err := item.run(); err != nil {
					return m, nil, "Error: " + err.Error()
				}
				return m, nil, item.label + " — done"
			}
		}
	}
	return m, nil, ""
}

func (m actionsModel) View() string {
	var b string
	b += actionMenuStyle.Render(func() string {
		s := titleStyle.Render(m.project.Name) + "\n\n"
		for i, item := range m.items {
			if i == m.cursor {
				s += selectedItemStyle.Render("▸ "+item.label) + "\n"
			} else {
				s += itemStyle.Render("  "+item.label) + "\n"
			}
		}
		s += "\n" + helpStyle.Render("enter: run  esc: back")
		return s
	}())
	return b
}

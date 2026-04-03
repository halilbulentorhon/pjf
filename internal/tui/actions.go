package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/halilbulentorhon/pjf/internal/project"
	"github.com/halilbulentorhon/pjf/internal/service"
)

type actionResult struct {
	status string
	action string
}

type actionItem struct {
	label  string
	action string
	run    func() error
}

type actionsModel struct {
	project       project.Project
	items         []actionItem
	cursor        int
	confirming    bool
	confirmLabel  string
	confirmFunc   func() error
	confirmAction string
}

func newActionsModel() actionsModel {
	return actionsModel{}
}

func newActionsModelForProject(p project.Project, svc *service.ProjectService, hidden bool) actionsModel {
	items := []actionItem{
		{label: "Open in IDE", action: "ide-submenu"},
		{label: "Open in Terminal", action: "open", run: func() error { return svc.OpenTerminal(p) }},
		{label: "Run Command", action: "cmd-submenu"},
		{label: "Copy Path", action: "copy", run: func() error { return svc.CopyPath(p) }},
	}

	if hidden {
		items = append(items, actionItem{
			label:  "Unhide",
			action: "unhide",
			run: func() error {
				svc.UnhideProject(p)
				return nil
			},
		})
	} else {
		items = append(items, actionItem{
			label:  "Hide from List",
			action: "hide",
			run: func() error {
				svc.HideProject(p)
				return nil
			},
		})
	}

	items = append(items, actionItem{
		label:  "Delete Project",
		action: "delete",
		run: func() error {
			return svc.DeleteProject(p)
		},
	})

	return actionsModel{
		project: p,
		items:   items,
	}
}

func (m actionsModel) Update(msg tea.Msg) (actionsModel, tea.Cmd, actionResult) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.confirming {
			switch msg.String() {
			case "y":
				if err := m.confirmFunc(); err != nil {
					return m, nil, actionResult{status: "Error: " + err.Error()}
				}
				m.confirming = false
				return m, nil, actionResult{status: m.confirmLabel + " — done", action: m.confirmAction}
			case "n", "esc":
				m.confirming = false
				return m, nil, actionResult{}
			}
			return m, nil, actionResult{}
		}

		key := msg.String()
		if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
			idx := int(key[0]-'1')
			if idx < len(m.items) {
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
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.items) {
				item := m.items[m.cursor]
				switch item.action {
				case "ide-submenu":
					return m, nil, actionResult{action: "ide-submenu"}
				case "cmd-submenu":
					return m, nil, actionResult{action: "cmd-submenu"}
				case "hide":
					m.confirming = true
					m.confirmLabel = item.label
					m.confirmFunc = item.run
					m.confirmAction = "hide"
					return m, nil, actionResult{}
				case "delete":
					m.confirming = true
					m.confirmLabel = item.label
					m.confirmFunc = item.run
					m.confirmAction = "delete"
					return m, nil, actionResult{}
				default:
					if err := item.run(); err != nil {
						return m, nil, actionResult{status: "Error: " + err.Error()}
					}
					return m, nil, actionResult{status: item.label + " — done", action: item.action}
				}
			}
		}
	}
	return m, nil, actionResult{}
}

func (m actionsModel) View() string {
	var b string
	b += actionMenuStyle.Render(func() string {
		title := m.project.Name
		s := titleStyle.Render(title) + "\n\n"

		if m.confirming {
			switch m.confirmAction {
			case "hide":
				s += confirmStyle.Render(fmt.Sprintf("Hide %q from list? (y/n)", m.project.Name))
			case "delete":
				s += confirmStyle.Render(fmt.Sprintf("Delete %q?\nThis cannot be undone. (y/n)", m.project.Path))
			}
			return s
		}

		for i, item := range m.items {
			num := fmt.Sprintf("%d. ", i+1)
			if i == m.cursor {
				s += selectedItemStyle.Render("▸ "+num+item.label) + "\n"
			} else {
				s += itemStyle.Render("  "+num+item.label) + "\n"
			}
		}
		s += "\n" + helpStyle.Render("1-9: select  enter: run  esc: back")
		return s
	}())
	return b
}

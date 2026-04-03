package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/halilbulentorhon/pjf/internal/ide"
	"github.com/halilbulentorhon/pjf/internal/project"
	"github.com/halilbulentorhon/pjf/internal/service"
)

type ideSubmenuResult struct {
	status string
	action string
}

type ideSubmenuModel struct {
	project    project.Project
	ides       []ide.IDE
	defaultIDE string
	cursor     int
	lastIDEIdx int
	svc        *service.ProjectService
}

func newIDESubmenuModel(p project.Project, svc *service.ProjectService) ideSubmenuModel {
	ides := svc.DetectedIDEs()
	defaultSlug := ""
	if resolved, ok := svc.ResolveIDE(p); ok {
		defaultSlug = resolved.Slug
	}
	return ideSubmenuModel{
		project:    p,
		ides:       ides,
		defaultIDE: defaultSlug,
		svc:        svc,
	}
}

func (m ideSubmenuModel) totalItems() int {
	return len(m.ides) + 1
}

func (m ideSubmenuModel) Update(msg tea.Msg) (ideSubmenuModel, tea.Cmd, ideSubmenuResult) {
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
			if m.cursor < len(m.ides) {
				selected := m.ides[m.cursor]
				if err := m.svc.OpenIDE(m.project, selected); err != nil {
					return m, nil, ideSubmenuResult{status: "Error: " + err.Error()}
				}
				return m, nil, ideSubmenuResult{
					status: "Open in " + selected.Name + " — done",
					action: "open-ide",
				}
			}
			if m.cursor == len(m.ides) && len(m.ides) > 0 {
				selected := m.ides[m.lastIDEIdx]
				m.svc.SetProjectIDE(m.project, selected.Slug)
				return m, nil, ideSubmenuResult{
					status: "Default IDE set: " + selected.Name,
					action: "set-default-ide",
				}
			}
		}
	}
	if m.cursor < len(m.ides) {
		m.lastIDEIdx = m.cursor
	}
	return m, nil, ideSubmenuResult{}
}

func (m ideSubmenuModel) View() string {
	return actionMenuStyle.Render(func() string {
		s := titleStyle.Render("Open in IDE") + "\n\n"

		if len(m.ides) == 0 {
			s += helpStyle.Render("  No IDEs found") + "\n"
			s += "\n" + helpStyle.Render("esc: back")
			return s
		}

		for i, ide := range m.ides {
			label := ide.Name
			if ide.Slug == m.defaultIDE {
				label += " " + defaultMarkerStyle.Render("★")
			}
			if i == m.cursor {
				s += selectedItemStyle.Render("▸ "+label) + "\n"
			} else {
				s += itemStyle.Render("  "+label) + "\n"
			}
		}

		s += separatorStyle.Render("  ──────────────") + "\n"

		setDefaultLabel := "Set as default for this project"
		if m.cursor == len(m.ides) {
			s += selectedItemStyle.Render("▸ "+setDefaultLabel) + "\n"
		} else {
			s += itemStyle.Render("  "+setDefaultLabel) + "\n"
		}

		s += "\n" + helpStyle.Render("enter: select  esc: back")
		return s
	}())
}

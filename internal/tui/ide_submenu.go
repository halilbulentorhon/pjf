package tui

import (
	"fmt"

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
	svc        *service.ProjectService
	pickMode   bool
	PickedSlug string
}

func newIDEPickerModel(p project.Project, svc *service.ProjectService) ideSubmenuModel {
	m := newIDESubmenuModel(p, svc)
	m.pickMode = true
	return m
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

func (m ideSubmenuModel) Update(msg tea.Msg) (ideSubmenuModel, tea.Cmd, ideSubmenuResult) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
			idx := int(key[0] - '1')
			if idx < len(m.ides) {
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
			if m.cursor < len(m.ides)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.ides) {
				selected := m.ides[m.cursor]
				if m.pickMode {
					m.PickedSlug = selected.Slug
					return m, nil, ideSubmenuResult{
						status: selected.Name,
						action: "pick",
					}
				}
				if err := m.svc.OpenIDE(m.project, selected); err != nil {
					return m, nil, ideSubmenuResult{status: "Error: " + err.Error()}
				}
				return m, nil, ideSubmenuResult{
					status: "Open in " + selected.Name + " — done",
					action: "open-ide",
				}
			}
		case "d":
			if m.cursor < len(m.ides) {
				selected := m.ides[m.cursor]
				m.svc.SetProjectIDE(m.project, selected.Slug)
				m.defaultIDE = selected.Slug
				return m, nil, ideSubmenuResult{
					status: "Default IDE set: " + selected.Name,
					action: "set-default-ide",
				}
			}
		}
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
			num := fmt.Sprintf("%d. ", i+1)
			label := ide.Name
			if ide.Slug == m.defaultIDE {
				label += " " + defaultMarkerStyle.Render("★")
			}
			if i == m.cursor {
				s += selectedItemStyle.Render("▸ "+num+label) + "\n"
			} else {
				s += itemStyle.Render("  "+num+label) + "\n"
			}
		}

		if m.pickMode {
			s += "\n" + helpStyle.Render("1-9: select  enter: pick  esc: cancel")
		} else {
			s += "\n" + helpStyle.Render("1-9: select  enter: open  d: set default  esc: back")
		}
		return s
	}())
}

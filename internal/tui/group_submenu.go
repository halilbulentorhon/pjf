package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/halilbulentorhon/pjf/internal/project"
	"github.com/halilbulentorhon/pjf/internal/service"
)

type groupSubmenuResult struct {
	action string
}

type groupSubmenuModel struct {
	project      project.Project
	svc          *service.ProjectService
	groups       []string
	currentGroup string
	cursor       int
	adding       bool
	addInput     settingsInputModel
}

func newGroupSubmenuModel(p project.Project, svc *service.ProjectService) groupSubmenuModel {
	var groups []string
	for _, g := range svc.Cfg.Groups {
		groups = append(groups, g.Name)
	}
	return groupSubmenuModel{
		project:      p,
		svc:          svc,
		groups:       groups,
		currentGroup: svc.ProjectGroup(p),
	}
}

func (m groupSubmenuModel) totalItems() int {
	count := len(m.groups) + 1
	if m.currentGroup != "" {
		count++
	}
	return count
}

func (m groupSubmenuModel) Update(msg tea.Msg) (groupSubmenuModel, tea.Cmd, groupSubmenuResult) {
	if m.adding {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				m.adding = false
				return m, nil, groupSubmenuResult{}
			}
		}
		inp, cmd, result := m.addInput.Update(msg)
		m.addInput = inp
		if result.cancelled {
			m.adding = false
			return m, nil, groupSubmenuResult{}
		}
		if result.value != "" {
			m.svc.AddGroup(result.value)
			m.svc.SetProjectGroup(m.project, result.value)
			m.adding = false
			return m, nil, groupSubmenuResult{action: "group-changed"}
		}
		return m, cmd, groupSubmenuResult{}
	}

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
			if m.cursor < len(m.groups) {
				selected := m.groups[m.cursor]
				m.svc.SetProjectGroup(m.project, selected)
				return m, nil, groupSubmenuResult{action: "group-changed"}
			}
			newGroupIdx := len(m.groups)
			if m.cursor == newGroupIdx {
				m.adding = true
				m.addInput = newSettingsInput("New Group Name", "name...")
				return m, textinput.Blink, groupSubmenuResult{}
			}
			if m.currentGroup != "" && m.cursor == newGroupIdx+1 {
				m.svc.RemoveProjectGroup(m.project)
				return m, nil, groupSubmenuResult{action: "group-changed"}
			}
		}
	}
	return m, nil, groupSubmenuResult{}
}

func (m groupSubmenuModel) View() string {
	if m.adding {
		return m.addInput.View()
	}

	return actionMenuStyle.Render(func() string {
		s := titleStyle.Render("Add to Group") + "\n\n"

		for i, g := range m.groups {
			label := g
			if g == m.currentGroup {
				label += " " + defaultMarkerStyle.Render("★")
			}
			if i == m.cursor {
				s += selectedItemStyle.Render("▸ "+label) + "\n"
			} else {
				s += itemStyle.Render("  "+label) + "\n"
			}
		}

		s += separatorStyle.Render("  ──────────────") + "\n"

		newLabel := "New Group..."
		newIdx := len(m.groups)
		if m.cursor == newIdx {
			s += selectedItemStyle.Render("▸ "+newLabel) + "\n"
		} else {
			s += itemStyle.Render("  "+newLabel) + "\n"
		}

		if m.currentGroup != "" {
			removeLabel := "Remove from Group"
			removeIdx := newIdx + 1
			if m.cursor == removeIdx {
				s += selectedItemStyle.Render("▸ "+removeLabel) + "\n"
			} else {
				s += itemStyle.Render("  "+removeLabel) + "\n"
			}
		}

		s += "\n" + helpStyle.Render("enter: select  esc: back")
		return s
	}())
}

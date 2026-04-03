package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/halilbulentorhon/pjf/internal/config"
	"github.com/halilbulentorhon/pjf/internal/project"
	"github.com/halilbulentorhon/pjf/internal/service"
)

type projectSettingsResult struct {
	changed bool
}

type projectSettingsModel struct {
	project   project.Project
	svc       *service.ProjectService
	commands  []config.CommandDef
	ideSlug   string
	cursor    int
	picking   bool
	idePicker ideSubmenuModel
	adding    bool
	addStep   int
	addName   string
	addInput  settingsInputModel
}

func newProjectSettingsModel(p project.Project, svc *service.ProjectService) projectSettingsModel {
	var commands []config.CommandDef
	for _, pc := range svc.Cfg.ProjectCommands {
		if pc.Path == p.Path {
			commands = pc.Commands
			break
		}
	}
	ideSlug := ""
	if svc.Cfg.ProjectIDEs != nil {
		ideSlug = svc.Cfg.ProjectIDEs[p.Path]
	}
	return projectSettingsModel{
		project:  p,
		svc:      svc,
		commands: commands,
		ideSlug:  ideSlug,
	}
}

func (m projectSettingsModel) totalItems() int {
	return 1 + len(m.commands)
}

func (m *projectSettingsModel) refreshCommands() {
	m.commands = nil
	for _, pc := range m.svc.Cfg.ProjectCommands {
		if pc.Path == m.project.Path {
			m.commands = pc.Commands
			break
		}
	}
}

func (m projectSettingsModel) Update(msg tea.Msg) (projectSettingsModel, tea.Cmd, projectSettingsResult) {
	if m.picking {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				m.picking = false
				return m, nil, projectSettingsResult{}
			}
		}
		sub, cmd, result := m.idePicker.Update(msg)
		m.idePicker = sub
		if result.action == "pick" {
			m.svc.SetProjectIDE(m.project, m.idePicker.PickedSlug)
			m.ideSlug = m.idePicker.PickedSlug
			m.picking = false
			return m, nil, projectSettingsResult{changed: true}
		}
		return m, cmd, projectSettingsResult{}
	}

	if m.adding {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				m.adding = false
				m.addStep = 0
				return m, nil, projectSettingsResult{}
			}
		}
		inp, cmd, result := m.addInput.Update(msg)
		m.addInput = inp
		if result.cancelled {
			m.adding = false
			m.addStep = 0
			return m, nil, projectSettingsResult{}
		}
		if result.value != "" {
			if m.addStep == 0 {
				m.addName = result.value
				m.addStep = 1
				m.addInput = newSettingsInput("Command", "command...")
				return m, textinput.Blink, projectSettingsResult{}
			}
			found := false
			for i, pc := range m.svc.Cfg.ProjectCommands {
				if pc.Path == m.project.Path {
					m.svc.Cfg.ProjectCommands[i].Commands = append(m.svc.Cfg.ProjectCommands[i].Commands, config.CommandDef{
						Name:    m.addName,
						Command: result.value,
					})
					found = true
					break
				}
			}
			if !found {
				m.svc.Cfg.ProjectCommands = append(m.svc.Cfg.ProjectCommands, config.ProjectCommandSet{
					Path:     m.project.Path,
					Commands: []config.CommandDef{{Name: m.addName, Command: result.value}},
				})
			}
			m.refreshCommands()
			m.adding = false
			m.addStep = 0
			return m, nil, projectSettingsResult{changed: true}
		}
		return m, cmd, projectSettingsResult{}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < m.totalItems()-1 {
				m.cursor++
			}
		case "e":
			if m.cursor == 0 {
				m.picking = true
				m.idePicker = newIDEPickerModel(m.project, m.svc)
				return m, nil, projectSettingsResult{}
			}
		case "d":
			if m.cursor == 0 && m.ideSlug != "" {
				m.svc.RemoveProjectIDE(m.project)
				m.ideSlug = ""
				return m, nil, projectSettingsResult{changed: true}
			}
			if m.cursor > 0 {
				idx := m.cursor - 1
				m.svc.RemoveSavedCommand(m.project, idx)
				m.refreshCommands()
				if m.cursor >= m.totalItems() && m.cursor > 0 {
					m.cursor--
				}
				return m, nil, projectSettingsResult{changed: true}
			}
		case "a":
			m.adding = true
			m.addStep = 0
			m.addInput = newSettingsInput("Command Name", "name...")
			return m, textinput.Blink, projectSettingsResult{}
		}
	}
	return m, nil, projectSettingsResult{}
}

func (m projectSettingsModel) View() string {
	if m.picking {
		return m.idePicker.View()
	}
	if m.adding {
		return m.addInput.View()
	}

	return actionMenuStyle.Render(func() string {
		s := titleStyle.Render("Project Settings — "+m.project.Name) + "\n\n"

		s += titleStyle.Render("Default IDE") + "\n"
		ideLabel := "(no override)"
		if m.ideSlug != "" {
			ideLabel = m.ideSlug + " ★"
		}
		if m.cursor == 0 {
			s += selectedItemStyle.Render("▸ "+ideLabel) + "\n"
		} else {
			s += itemStyle.Render("  "+ideLabel) + "\n"
		}
		s += "\n"

		s += titleStyle.Render("Saved Commands") + "\n"
		if len(m.commands) == 0 {
			s += helpStyle.Render("  (none)") + "\n"
		}
		for i, cmd := range m.commands {
			label := fmt.Sprintf("%s — %s", cmd.Name, helpStyle.Render(cmd.Command))
			if i+1 == m.cursor {
				s += selectedItemStyle.Render("▸ "+label) + "\n"
			} else {
				s += itemStyle.Render("  "+label) + "\n"
			}
		}

		s += "\n" + helpStyle.Render("e: change IDE  d: delete  a: add command  esc: back")
		return s
	}())
}

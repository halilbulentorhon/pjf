package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halilbulentorhon/pjf/internal/ide"
	"github.com/halilbulentorhon/pjf/internal/project"
	"github.com/halilbulentorhon/pjf/internal/service"
)

type appState int

const (
	stateWizard appState = iota
	stateScanning
	stateList
	stateActions
	stateHelp
	stateIDESubmenu
	stateCommandSubmenu
	stateCommandInput
	stateOutput
	stateGlobalSettings
	stateProjectSettings
	stateGroupSubmenu
)

type scanCompleteMsg struct {
	projects     []project.Project
	err          error
	needsRefresh bool
}

type refreshCompleteMsg struct {
	projects []project.Project
	err      error
}

type wizardCompleteMsg struct{}

type statusMsg string

type commandDoneMsg struct {
	title  string
	output string
	err    error
}

type Model struct {
	state      appState
	prevState  appState
	service    *service.ProjectService
	configPath string
	wizard     wizardModel
	list       listModel
	actions    actionsModel
	help       helpModel
	ideSubmenu  ideSubmenuModel
	cmdSubmenu  cmdSubmenuModel
	cmdInput    cmdInputModel
	output          outputModel
	globalSettings  globalSettingsModel
	projectSettings projectSettingsModel
	groupSubmenu    groupSubmenuModel
	status          string
	width      int
	height     int
}

func New(svc *service.ProjectService, configPath string, isFirstRun bool) Model {
	m := Model{
		service:    svc,
		configPath: configPath,
		wizard:     newWizardModel(),
		list: newListModel(svc.IsHidden, func(p project.Project) string {
			if resolved, ok := svc.ResolveIDE(p); ok {
				return resolved.Name
			}
			return ""
		}, svc.GroupedProjects),
		actions:    newActionsModel(),
		help:       newHelpModel(),
	}
	if isFirstRun {
		m.state = stateWizard
	} else {
		m.state = stateScanning
	}
	return m
}

func (m Model) Init() tea.Cmd {
	if m.state == stateWizard {
		return m.wizard.Init()
	}
	return m.loadFromCacheCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.setSize(msg.Width, msg.Height)
		m.output.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case scanCompleteMsg:
		if msg.err != nil {
			m.status = "Scan error: " + msg.err.Error()
			return m, nil
		}
		m.service.SetDetectedIDEs(ide.DetectAll())
		m.list.setProjects(msg.projects)
		m.state = stateList
		m.status = ""
		if msg.needsRefresh {
			return m, m.refreshCmd()
		}
		return m, nil

	case refreshCompleteMsg:
		if msg.err == nil && msg.projects != nil {
			m.list.setProjects(msg.projects)
			m.status = "List updated"
		}
		return m, nil

	case wizardCompleteMsg:
		m.state = stateScanning
		return m, m.scanCmd()

	case statusMsg:
		m.status = string(msg)
		return m, nil

	case commandDoneMsg:
		if m.state == stateOutput {
			output := msg.output
			if msg.err != nil {
				output += "\n" + msg.err.Error()
			}
			m.output.setOutput(output)
		}
		return m, nil
	}

	switch m.state {
	case stateWizard:
		return m.updateWizard(msg)
	case stateScanning:
		return m, nil
	case stateList:
		return m.updateList(msg)
	case stateActions:
		return m.updateActions(msg)
	case stateHelp:
		return m.updateHelp(msg)
	case stateIDESubmenu:
		return m.updateIDESubmenu(msg)
	case stateCommandSubmenu:
		return m.updateCommandSubmenu(msg)
	case stateCommandInput:
		return m.updateCommandInput(msg)
	case stateOutput:
		return m.updateOutput(msg)
	case stateGlobalSettings:
		return m.updateGlobalSettings(msg)
	case stateProjectSettings:
		return m.updateProjectSettings(msg)
	case stateGroupSubmenu:
		return m.updateGroupSubmenu(msg)
	}
	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case stateWizard:
		return m.wizard.View()
	case stateScanning:
		return m.viewScanning()
	case stateList:
		return m.list.View(m.width, m.height, m.status)
	case stateActions:
		return m.actions.View()
	case stateHelp:
		return m.help.View(m.width, m.height)
	case stateIDESubmenu:
		return m.ideSubmenu.View()
	case stateCommandSubmenu:
		return m.cmdSubmenu.View()
	case stateCommandInput:
		return m.cmdInput.View()
	case stateOutput:
		return m.output.View()
	case stateGlobalSettings:
		return m.globalSettings.View()
	case stateProjectSettings:
		return m.projectSettings.View()
	case stateGroupSubmenu:
		return m.groupSubmenu.View()
	}
	return ""
}

func (m Model) viewScanning() string {
	return wizardTextStyle.Render("Scanning for projects...")
}

func (m Model) loadFromCacheCmd() tea.Cmd {
	return func() tea.Msg {
		projects, fromCache, needsRefresh, err := m.service.LoadOrScan(context.Background())
		if err != nil {
			return scanCompleteMsg{err: err}
		}
		if !fromCache {
			projects, err := m.service.Scan(context.Background(), nil)
			return scanCompleteMsg{projects: projects, err: err}
		}
		return scanCompleteMsg{projects: projects, needsRefresh: needsRefresh}
	}
}

func (m Model) scanCmd() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.service.Scan(context.Background(), nil)
		return scanCompleteMsg{projects: projects, err: err}
	}
}

func (m Model) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.service.Refresh(context.Background())
		return refreshCompleteMsg{projects: projects, err: err}
	}
}

func (m *Model) updateWizard(msg tea.Msg) (tea.Model, tea.Cmd) {
	wiz, cmd, done := m.wizard.Update(msg)
	m.wizard = wiz
	if done {
		m.service.Cfg.ScanDirs = m.wizard.dirs
		m.service.SaveConfig(m.configPath)
		return m, func() tea.Msg { return wizardCompleteMsg{} }
	}
	return m, cmd
}

func (m *Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.confirmingHide {
			switch msg.String() {
			case "y":
				m.service.HideProject(m.list.confirmProject)
				m.service.SaveConfig(m.configPath)
				m.list.confirmingHide = false
				m.list.rebuildSections()
				m.status = "Hidden: " + m.list.confirmProject.Name
			default:
				m.list.confirmingHide = false
				m.status = ""
			}
			return m, nil
		}
		if !m.list.searchFocused {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "?":
				m.prevState = stateList
				m.state = stateHelp
				return m, nil
			case "r":
				m.status = "Rescanning..."
				m.state = stateScanning
				return m, m.scanCmd()
			case "h":
				if p, ok := m.list.selected(); ok {
					m.list.confirmingHide = true
					m.list.confirmProject = p
					m.status = fmt.Sprintf("Hide %q? (y/n)", p.Name)
					return m, nil
				}
			case "ctrl+h":
				m.list.showHidden = !m.list.showHidden
				m.list.rebuildSections()
				if m.list.showHidden {
					hasHidden := false
					for _, p := range m.list.projects {
						if m.service.IsHidden(p) {
							hasHidden = true
							break
						}
					}
					if !hasHidden {
						m.status = "No hidden projects"
						m.list.showHidden = false
					} else {
						m.status = "Showing hidden projects"
					}
				} else {
					m.status = ""
				}
				return m, nil
			case "t":
				if p, ok := m.list.selected(); ok {
					if err := m.service.OpenTerminal(p); err != nil {
						m.status = "Error: " + err.Error()
					} else {
						m.status = "Open in Terminal — done"
					}
					return m, nil
				}
			case "s":
				m.globalSettings = newGlobalSettingsModel(m.service)
				m.state = stateGlobalSettings
				return m, nil
			case "o":
				if p, ok := m.list.selected(); ok {
					resolved, ok := m.service.ResolveIDE(p)
					if !ok {
						m.status = "No IDE configured"
						return m, nil
					}
					if err := m.service.OpenIDE(p, resolved); err != nil {
						m.status = "Error: " + err.Error()
					} else {
						m.status = "Open in " + resolved.Name + " — done"
					}
					return m, nil
				}
			case "m":
				if p, ok := m.list.selected(); ok {
					m.groupSubmenu = newGroupSubmenuModel(p, m.service)
					m.state = stateGroupSubmenu
					return m, nil
				}
			case "u":
				if m.list.cursor >= 0 && m.list.cursor < len(m.list.flatItems) {
					item := m.list.flatItems[m.list.cursor]
					if item.isHeader && item.groupIndex >= 0 {
						m.service.MoveGroupUp(item.groupName)
						m.service.SaveConfig(m.configPath)
						m.list.rebuildSections()
						if m.list.cursor > 0 {
							for i := m.list.cursor - 1; i >= 0; i-- {
								if m.list.flatItems[i].isHeader && m.list.flatItems[i].groupName == item.groupName {
									m.list.cursor = i
									break
								}
							}
						}
					}
				}
				return m, nil
			case "d":
				if m.list.cursor >= 0 && m.list.cursor < len(m.list.flatItems) {
					item := m.list.flatItems[m.list.cursor]
					if item.isHeader && item.groupIndex >= 0 {
						m.service.MoveGroupDown(item.groupName)
						m.service.SaveConfig(m.configPath)
						m.list.rebuildSections()
						for i := 0; i < len(m.list.flatItems); i++ {
							if m.list.flatItems[i].isHeader && m.list.flatItems[i].groupName == item.groupName {
								m.list.cursor = i
								break
							}
						}
					}
				}
				return m, nil
			case "enter":
				if p, ok := m.list.selected(); ok {
					hidden := m.service.IsHidden(p)
					m.actions = newActionsModelForProject(p, m.service, hidden)
					m.state = stateActions
					return m, nil
				}
			}
		}
	}
	lst, cmd, collapseToggle := m.list.Update(msg)
	m.list = lst
	if collapseToggle {
		m.list.toggleCollapse(m.service)
		m.service.SaveConfig(m.configPath)
	}
	return m, cmd
}

func (m *Model) updateActions(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" && !m.actions.confirming {
			m.state = stateList
			return m, nil
		}
	}
	act, cmd, result := m.actions.Update(msg)
	m.actions = act
	if result.action == "ide-submenu" {
		m.ideSubmenu = newIDESubmenuModel(m.actions.project, m.service)
		m.state = stateIDESubmenu
		return m, cmd
	}
	if result.action == "cmd-submenu" {
		m.cmdSubmenu = newCmdSubmenuModel(m.actions.project, m.service)
		m.state = stateCommandSubmenu
		return m, cmd
	}
	if result.action == "project-settings" {
		m.projectSettings = newProjectSettingsModel(m.actions.project, m.service)
		m.state = stateProjectSettings
		return m, cmd
	}
	if result.action == "group-submenu" {
		m.groupSubmenu = newGroupSubmenuModel(m.actions.project, m.service)
		m.state = stateGroupSubmenu
		return m, cmd
	}
	if result.status != "" {
		m.status = result.status
		switch result.action {
		case "hide":
			m.service.SaveConfig(m.configPath)
			m.list.rebuildSections()
			m.state = stateList
		case "unhide":
			m.service.SaveConfig(m.configPath)
			m.list.rebuildSections()
			m.state = stateList
		case "delete":
			m.service.SaveConfig(m.configPath)
			m.service.RemoveFromCache(m.actions.project)
			m.list.removeProject(m.actions.project.Path)
			m.state = stateList
		default:
			m.state = stateList
		}
	}
	return m, cmd
}

func (m *Model) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "?" {
			m.state = m.prevState
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) updateIDESubmenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.state = stateActions
			return m, nil
		}
	}
	sub, cmd, result := m.ideSubmenu.Update(msg)
	m.ideSubmenu = sub
	if result.status != "" {
		m.status = result.status
		switch result.action {
		case "set-default-ide":
			m.service.SaveConfig(m.configPath)
		default:
			m.state = stateList
		}
	}
	return m, cmd
}

func (m *Model) updateCommandSubmenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.state = stateActions
			return m, nil
		}
	}
	sub, cmd, result := m.cmdSubmenu.Update(msg)
	m.cmdSubmenu = sub
	switch result.action {
	case "run":
		p := m.cmdSubmenu.project
		command := result.command
		m.output = newOutputModel(command, m.height)
		m.state = stateOutput
		return m, func() tea.Msg {
			output, err := m.service.RunCommand(p, command)
			return commandDoneMsg{title: command, output: output, err: err}
		}
	case "custom-input":
		m.cmdInput = newCmdInputModel()
		m.state = stateCommandInput
		return m, textinput.Blink
	}
	return m, cmd
}

func (m *Model) updateCommandInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.state = stateCommandSubmenu
			return m, nil
		}
	}
	inp, cmd, result := m.cmdInput.Update(msg)
	m.cmdInput = inp
	if result.command != "" {
		p := m.cmdSubmenu.project
		command := result.command
		if result.save {
			m.service.SaveCommand(p, command)
			m.service.SaveConfig(m.configPath)
		}
		m.output = newOutputModel(command, m.height)
		m.state = stateOutput
		return m, func() tea.Msg {
			output, err := m.service.RunCommand(p, command)
			return commandDoneMsg{title: command, output: output, err: err}
		}
	}
	return m, cmd
}

func (m *Model) updateGlobalSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" && !m.globalSettings.Inputting() && !m.globalSettings.Picking() {
			needsRescan := m.globalSettings.NeedsRescan()
			m.service.SaveConfig(m.configPath)
			m.state = stateList
			if needsRescan {
				m.status = "Rescanning..."
				m.state = stateScanning
				return m, m.scanCmd()
			}
			return m, nil
		}
	}
	gs, cmd, result := m.globalSettings.Update(msg)
	m.globalSettings = gs
	if result.changed {
		m.service.SaveConfig(m.configPath)
	}
	return m, cmd
}

func (m *Model) updateProjectSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" && !m.projectSettings.picking && !m.projectSettings.adding {
			m.state = stateActions
			return m, nil
		}
	}
	ps, cmd, result := m.projectSettings.Update(msg)
	m.projectSettings = ps
	if result.changed {
		m.service.SaveConfig(m.configPath)
	}
	return m, cmd
}

func (m *Model) updateGroupSubmenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" && !m.groupSubmenu.adding {
			m.state = stateActions
			return m, nil
		}
	}
	sub, cmd, result := m.groupSubmenu.Update(msg)
	m.groupSubmenu = sub
	if result.action == "group-changed" {
		m.service.SaveConfig(m.configPath)
		m.list.rebuildSections()
		m.state = stateList
		m.status = "Group updated"
	}
	return m, cmd
}

func (m *Model) updateOutput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.state = stateActions
			return m, nil
		}
	}
	out, cmd := m.output.Update(msg)
	m.output = out
	return m, cmd
}

func Run(svc *service.ProjectService, configPath string, isFirstRun bool) error {
	m := New(svc, configPath, isFirstRun)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

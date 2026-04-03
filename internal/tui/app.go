package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
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

type Model struct {
	state      appState
	prevState  appState
	service    *service.ProjectService
	configPath string
	wizard     wizardModel
	list       listModel
	actions    actionsModel
	help       helpModel
	status     string
	width      int
	height     int
}

func New(svc *service.ProjectService, configPath string, isFirstRun bool) Model {
	m := Model{
		service:    svc,
		configPath: configPath,
		wizard:     newWizardModel(),
		list:       newListModel(),
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
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case scanCompleteMsg:
		if msg.err != nil {
			m.status = "Tarama hatası: " + msg.err.Error()
			return m, nil
		}
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
			m.status = "Liste güncellendi"
		}
		return m, nil

	case wizardCompleteMsg:
		m.state = stateScanning
		return m, m.scanCmd()

	case statusMsg:
		m.status = string(msg)
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
	}
	return ""
}

func (m Model) viewScanning() string {
	return wizardTextStyle.Render("Projeler taranıyor...")
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
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "?":
			m.prevState = stateList
			m.state = stateHelp
			return m, nil
		case "r":
			m.status = "Yeniden taranıyor..."
			m.state = stateScanning
			return m, m.scanCmd()
		case "enter":
			if p, ok := m.list.selected(); ok {
				m.actions = newActionsModelForProject(p, m.service)
				m.state = stateActions
				return m, nil
			}
		}
	}
	lst, cmd := m.list.Update(msg)
	m.list = lst
	return m, cmd
}

func (m *Model) updateActions(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.state = stateList
			return m, nil
		}
	}
	act, cmd, status := m.actions.Update(msg)
	m.actions = act
	if status != "" {
		m.status = status
		m.state = stateList
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

func Run(svc *service.ProjectService, configPath string, isFirstRun bool) error {
	m := New(svc, configPath, isFirstRun)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

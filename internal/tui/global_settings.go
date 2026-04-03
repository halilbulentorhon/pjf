package tui

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/halilbulentorhon/pjf/internal/project"
	"github.com/halilbulentorhon/pjf/internal/service"
)

type settingsTab int

const (
	tabScan settingsTab = iota
	tabIDEs
	tabCommands
	tabOther
)

var tabNames = []string{"Scan", "IDEs", "Commands", "Other"}

type pendingOp int

const (
	opNone pendingOp = iota
	opAddScanDir
	opAddExclude
	opAddIDEType
	opAddIDESlug
	opAddCmdName
	opAddCmdCommand
	opEditMaxDepth
	opEditCacheTTL
	opEditIDESlug
)

type globalSettingsResult struct {
	changed bool
}

type globalSettingsModel struct {
	svc             *service.ProjectService
	activeTab       settingsTab
	cursors         [4]int
	scanDirsChanged bool

	inputting bool
	input     settingsInputModel
	pending   pendingOp
	tempStr   string

	picking     bool
	idePicker   ideSubmenuModel
	editIDEType string
}

func newGlobalSettingsModel(svc *service.ProjectService) globalSettingsModel {
	return globalSettingsModel{svc: svc}
}

func (m globalSettingsModel) NeedsRescan() bool {
	return m.scanDirsChanged
}

func (m globalSettingsModel) Inputting() bool {
	return m.inputting
}

func (m globalSettingsModel) Picking() bool {
	return m.picking
}

func (m globalSettingsModel) ideKeys() []string {
	if m.svc.Cfg.DefaultIDEs == nil {
		return nil
	}
	keys := make([]string, 0, len(m.svc.Cfg.DefaultIDEs))
	for k := range m.svc.Cfg.DefaultIDEs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (m globalSettingsModel) tabItemCount() int {
	switch m.activeTab {
	case tabScan:
		return len(m.svc.Cfg.ScanDirs) + len(m.svc.Cfg.ExtraExcludes)
	case tabIDEs:
		return len(m.ideKeys())
	case tabCommands:
		return len(m.svc.Cfg.GlobalCommands)
	case tabOther:
		return 2
	}
	return 0
}

func (m globalSettingsModel) Update(msg tea.Msg) (globalSettingsModel, tea.Cmd, globalSettingsResult) {
	if m.inputting {
		inp, cmd, result := m.input.Update(msg)
		m.input = inp
		if result.cancelled {
			m.inputting = false
			m.pending = opNone
			return m, nil, globalSettingsResult{}
		}
		if result.value != "" {
			return m.handleInputComplete(result.value)
		}
		return m, cmd, globalSettingsResult{}
	}

	if m.picking {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				m.picking = false
				m.pending = opNone
				return m, nil, globalSettingsResult{}
			}
		}
		sub, cmd, result := m.idePicker.Update(msg)
		m.idePicker = sub
		if result.action == "pick" {
			return m.handlePickComplete(m.idePicker.PickedSlug)
		}
		return m, cmd, globalSettingsResult{}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if m.activeTab > 0 {
				m.activeTab--
			}
		case "right":
			if m.activeTab < tabOther {
				m.activeTab++
			}
		case "up":
			if m.cursors[m.activeTab] > 0 {
				m.cursors[m.activeTab]--
			}
		case "down":
			max := m.tabItemCount() - 1
			if max < 0 {
				max = 0
			}
			if m.cursors[m.activeTab] < max {
				m.cursors[m.activeTab]++
			}
		case "a":
			return m.handleAdd()
		case "d":
			return m.handleDelete()
		case "e":
			return m.handleEdit()
		}
	}
	return m, nil, globalSettingsResult{}
}

func (m globalSettingsModel) handleInputComplete(value string) (globalSettingsModel, tea.Cmd, globalSettingsResult) {
	m.inputting = false
	switch m.pending {
	case opAddScanDir:
		m.svc.AddScanDir(expandHome(value))
		m.scanDirsChanged = true
		m.pending = opNone
		return m, nil, globalSettingsResult{changed: true}
	case opAddExclude:
		m.svc.AddExclude(expandHome(value))
		m.scanDirsChanged = true
		m.pending = opNone
		return m, nil, globalSettingsResult{changed: true}
	case opAddIDEType:
		m.tempStr = value
		m.pending = opAddIDESlug
		m.picking = true
		m.idePicker = newIDEPickerModel(project.Project{}, m.svc)
		return m, nil, globalSettingsResult{}
	case opAddCmdName:
		m.tempStr = value
		m.pending = opAddCmdCommand
		m.inputting = true
		m.input = newSettingsInput("Command", "command...")
		return m, textinput.Blink, globalSettingsResult{}
	case opAddCmdCommand:
		m.svc.AddGlobalCommand(m.tempStr, value)
		m.pending = opNone
		return m, nil, globalSettingsResult{changed: true}
	case opEditMaxDepth:
		if n, err := strconv.Atoi(value); err == nil && n > 0 {
			m.svc.SetMaxDepth(n)
		}
		m.pending = opNone
		return m, nil, globalSettingsResult{changed: true}
	case opEditCacheTTL:
		if n, err := strconv.Atoi(value); err == nil && n > 0 {
			m.svc.SetCacheTTL(n)
		}
		m.pending = opNone
		return m, nil, globalSettingsResult{changed: true}
	}
	m.pending = opNone
	return m, nil, globalSettingsResult{}
}

func (m globalSettingsModel) handlePickComplete(slug string) (globalSettingsModel, tea.Cmd, globalSettingsResult) {
	m.picking = false
	switch m.pending {
	case opAddIDESlug:
		m.svc.SetDefaultIDE(m.tempStr, slug)
		m.pending = opNone
		return m, nil, globalSettingsResult{changed: true}
	case opEditIDESlug:
		m.svc.SetDefaultIDE(m.editIDEType, slug)
		m.pending = opNone
		return m, nil, globalSettingsResult{changed: true}
	}
	m.pending = opNone
	return m, nil, globalSettingsResult{}
}

func (m globalSettingsModel) handleAdd() (globalSettingsModel, tea.Cmd, globalSettingsResult) {
	switch m.activeTab {
	case tabScan:
		cursor := m.cursors[tabScan]
		if cursor < len(m.svc.Cfg.ScanDirs) {
			m.inputting = true
			m.pending = opAddScanDir
			m.input = newSettingsInput("Add Scan Directory", "~/path...")
		} else {
			m.inputting = true
			m.pending = opAddExclude
			m.input = newSettingsInput("Add Exclude", "path...")
		}
		return m, textinput.Blink, globalSettingsResult{}
	case tabIDEs:
		m.inputting = true
		m.pending = opAddIDEType
		m.input = newSettingsInput("Project Type", "go, node, java, _default...")
		return m, textinput.Blink, globalSettingsResult{}
	case tabCommands:
		m.inputting = true
		m.pending = opAddCmdName
		m.input = newSettingsInput("Command Name", "name...")
		return m, textinput.Blink, globalSettingsResult{}
	}
	return m, nil, globalSettingsResult{}
}

func (m globalSettingsModel) handleDelete() (globalSettingsModel, tea.Cmd, globalSettingsResult) {
	cursor := m.cursors[m.activeTab]
	switch m.activeTab {
	case tabScan:
		if cursor < len(m.svc.Cfg.ScanDirs) {
			if len(m.svc.Cfg.ScanDirs) > 1 {
				m.svc.RemoveScanDir(cursor)
				m.scanDirsChanged = true
				if m.cursors[tabScan] >= m.tabItemCount() && m.cursors[tabScan] > 0 {
					m.cursors[tabScan]--
				}
				return m, nil, globalSettingsResult{changed: true}
			}
		} else {
			idx := cursor - len(m.svc.Cfg.ScanDirs)
			m.svc.RemoveExclude(idx)
			m.scanDirsChanged = true
			if m.cursors[tabScan] >= m.tabItemCount() && m.cursors[tabScan] > 0 {
				m.cursors[tabScan]--
			}
			return m, nil, globalSettingsResult{changed: true}
		}
	case tabIDEs:
		keys := m.ideKeys()
		if cursor < len(keys) {
			m.svc.RemoveDefaultIDE(keys[cursor])
			if m.cursors[tabIDEs] >= m.tabItemCount() && m.cursors[tabIDEs] > 0 {
				m.cursors[tabIDEs]--
			}
			return m, nil, globalSettingsResult{changed: true}
		}
	case tabCommands:
		if cursor < len(m.svc.Cfg.GlobalCommands) {
			m.svc.RemoveGlobalCommand(cursor)
			if m.cursors[tabCommands] >= m.tabItemCount() && m.cursors[tabCommands] > 0 {
				m.cursors[tabCommands]--
			}
			return m, nil, globalSettingsResult{changed: true}
		}
	}
	return m, nil, globalSettingsResult{}
}

func (m globalSettingsModel) handleEdit() (globalSettingsModel, tea.Cmd, globalSettingsResult) {
	cursor := m.cursors[m.activeTab]
	switch m.activeTab {
	case tabIDEs:
		keys := m.ideKeys()
		if cursor < len(keys) {
			m.editIDEType = keys[cursor]
			m.pending = opEditIDESlug
			m.picking = true
			m.idePicker = newIDEPickerModel(project.Project{}, m.svc)
			return m, nil, globalSettingsResult{}
		}
	case tabOther:
		if cursor == 0 {
			m.inputting = true
			m.pending = opEditMaxDepth
			m.input = newSettingsInput("Max Depth", strconv.Itoa(m.svc.Cfg.MaxDepth))
			return m, textinput.Blink, globalSettingsResult{}
		}
		if cursor == 1 {
			m.inputting = true
			m.pending = opEditCacheTTL
			m.input = newSettingsInput("Cache TTL (hours)", strconv.Itoa(m.svc.Cfg.CacheTTL))
			return m, textinput.Blink, globalSettingsResult{}
		}
	}
	return m, nil, globalSettingsResult{}
}

func (m globalSettingsModel) View() string {
	if m.inputting {
		return m.input.View()
	}
	if m.picking {
		return m.idePicker.View()
	}

	return actionMenuStyle.Render(func() string {
		s := titleStyle.Render("Settings") + "\n\n"

		for i, name := range tabNames {
			if settingsTab(i) == m.activeTab {
				s += activeTabStyle.Render("["+name+"]") + "  "
			} else {
				s += inactiveTabStyle.Render(name) + "  "
			}
		}
		s += "\n\n"

		cursor := m.cursors[m.activeTab]

		switch m.activeTab {
		case tabScan:
			s += titleStyle.Render("Scan Directories") + "\n"
			for i, d := range m.svc.Cfg.ScanDirs {
				if i == cursor {
					s += selectedItemStyle.Render("▸ "+d) + "\n"
				} else {
					s += itemStyle.Render("  "+d) + "\n"
				}
			}
			s += "\n" + titleStyle.Render("Extra Excludes") + "\n"
			if len(m.svc.Cfg.ExtraExcludes) == 0 {
				s += helpStyle.Render("  (none)") + "\n"
			}
			for i, d := range m.svc.Cfg.ExtraExcludes {
				idx := len(m.svc.Cfg.ScanDirs) + i
				if idx == cursor {
					s += selectedItemStyle.Render("▸ "+d) + "\n"
				} else {
					s += itemStyle.Render("  "+d) + "\n"
				}
			}

		case tabIDEs:
			keys := m.ideKeys()
			if len(keys) == 0 {
				s += helpStyle.Render("  (no default IDEs configured)") + "\n"
			} else {
				for i, k := range keys {
					label := fmt.Sprintf("%s → %s", k, m.svc.Cfg.DefaultIDEs[k])
					if i == cursor {
						s += selectedItemStyle.Render("▸ "+label) + "\n"
					} else {
						s += itemStyle.Render("  "+label) + "\n"
					}
				}
			}

		case tabCommands:
			if len(m.svc.Cfg.GlobalCommands) == 0 {
				s += helpStyle.Render("  (no global commands)") + "\n"
			}
			for i, cmd := range m.svc.Cfg.GlobalCommands {
				label := fmt.Sprintf("%s — %s", cmd.Name, helpStyle.Render(cmd.Command))
				if i == cursor {
					s += selectedItemStyle.Render("▸ "+label) + "\n"
				} else {
					s += itemStyle.Render("  "+label) + "\n"
				}
			}

		case tabOther:
			items := []string{
				fmt.Sprintf("Max Depth: %d", m.svc.Cfg.MaxDepth),
				fmt.Sprintf("Cache TTL: %dh", m.svc.Cfg.CacheTTL),
			}
			for i, item := range items {
				if i == cursor {
					s += selectedItemStyle.Render("▸ "+item) + "\n"
				} else {
					s += itemStyle.Render("  "+item) + "\n"
				}
			}
		}

		s += "\n" + helpStyle.Render("←/→: tab  ↑/↓: navigate  a: add  d: delete  e: edit  esc: back")
		return s
	}())
}

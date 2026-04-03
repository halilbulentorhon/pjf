package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"

	"github.com/halilbulentorhon/pjf/internal/project"
	"github.com/halilbulentorhon/pjf/internal/service"
)

type listItem struct {
	isHeader     bool
	groupIndex   int
	groupName    string
	collapsed    bool
	projectCount int
	project      project.Project
}

type listModel struct {
	search         textinput.Model
	projects       []project.Project
	flatItems      []listItem
	cursor         int
	width          int
	height         int
	showHidden     bool
	searchFocused  bool
	otherCollapsed bool
	isHidden       func(project.Project) bool
	resolveIDEName func(project.Project) string
	grouper        func([]project.Project) []service.GroupedSection
}

func newListModel(isHidden func(project.Project) bool, resolveIDEName func(project.Project) string, grouper func([]project.Project) []service.GroupedSection) listModel {
	ti := textinput.New()
	ti.Placeholder = "search..."
	ti.Prompt = "> "
	ti.Focus()
	return listModel{
		search:         ti,
		searchFocused:  true,
		isHidden:       isHidden,
		resolveIDEName: resolveIDEName,
		grouper:        grouper,
	}
}

func (m *listModel) setProjects(projects []project.Project) {
	m.projects = projects
	m.rebuildSections()
}

func (m *listModel) setSize(w, h int) {
	m.width = w
	m.height = h
}

func (m listModel) selected() (project.Project, bool) {
	if m.cursor < 0 || m.cursor >= len(m.flatItems) {
		return project.Project{}, false
	}
	item := m.flatItems[m.cursor]
	if item.isHeader {
		return project.Project{}, false
	}
	return item.project, true
}

func (m *listModel) rebuildSections() {
	var visible []project.Project
	if m.showHidden {
		visible = m.projects
	} else {
		for _, p := range m.projects {
			if !m.isHidden(p) {
				visible = append(visible, p)
			}
		}
	}

	query := strings.TrimSpace(m.search.Value())
	if query != "" {
		matches := fuzzy.FindFrom(query, projectSource(visible))
		m.flatItems = nil
		for _, match := range matches {
			m.flatItems = append(m.flatItems, listItem{project: visible[match.Index]})
		}
		m.clampCursor()
		return
	}

	sections := m.grouper(visible)

	m.flatItems = nil
	for si, sec := range sections {
		gIdx := si
		if sec.IsOther {
			gIdx = -1
		}
		collapsed := sec.Collapsed
		if sec.IsOther {
			collapsed = m.otherCollapsed
		}
		m.flatItems = append(m.flatItems, listItem{
			isHeader:     true,
			groupIndex:   gIdx,
			groupName:    sec.Name,
			collapsed:    collapsed,
			projectCount: len(sec.Projects),
		})
		if !collapsed {
			for _, p := range sec.Projects {
				m.flatItems = append(m.flatItems, listItem{
					project:    p,
					groupIndex: gIdx,
					groupName:  sec.Name,
				})
			}
		}
	}
	m.clampCursor()
}

func (m *listModel) clampCursor() {
	if m.cursor >= len(m.flatItems) {
		m.cursor = len(m.flatItems) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *listModel) toggleCollapse(svc *service.ProjectService) {
	if m.cursor < 0 || m.cursor >= len(m.flatItems) {
		return
	}
	item := m.flatItems[m.cursor]
	if !item.isHeader {
		return
	}
	if item.groupIndex == -1 {
		m.otherCollapsed = !m.otherCollapsed
	} else {
		svc.SetGroupCollapsed(item.groupName, !item.collapsed)
	}
	m.rebuildSections()
}

func (m listModel) Update(msg tea.Msg) (listModel, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchFocused {
			switch msg.String() {
			case "down":
				m.searchFocused = false
				m.search.Blur()
				return m, nil, false
			case "esc":
				if m.search.Value() == "" {
					return m, tea.Quit, false
				}
				m.search.SetValue("")
				m.rebuildSections()
				return m, nil, false
			case "enter":
				return m, nil, false
			}
			oldVal := m.search.Value()
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			if m.search.Value() != oldVal {
				m.rebuildSections()
			}
			return m, cmd, false
		}

		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.searchFocused = true
				m.search.Focus()
				return m, textinput.Blink, false
			}
			return m, nil, false
		case "down":
			if m.cursor < len(m.flatItems)-1 {
				m.cursor++
			}
			return m, nil, false
		case "esc":
			m.searchFocused = true
			m.search.Focus()
			return m, textinput.Blink, false
		case "left":
			if m.cursor >= 0 && m.cursor < len(m.flatItems) {
				item := m.flatItems[m.cursor]
				if item.isHeader {
					if !item.collapsed {
						return m, nil, true
					}
					return m, nil, false
				}
				for i := m.cursor - 1; i >= 0; i-- {
					if m.flatItems[i].isHeader && m.flatItems[i].groupName == item.groupName {
						m.cursor = i
						break
					}
				}
			}
			return m, nil, false
		case "right":
			if m.cursor >= 0 && m.cursor < len(m.flatItems) {
				item := m.flatItems[m.cursor]
				if item.isHeader && item.collapsed {
					return m, nil, true
				}
			}
			return m, nil, false
		}
	}
	return m, nil, false
}

func (m listModel) View(width, height int, status string) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("ProjectFinder") + "\n")
	b.WriteString(m.search.View() + "\n\n")

	maxVisible := height - 7
	if maxVisible < 1 {
		maxVisible = 10
	}

	home, _ := os.UserHomeDir()

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.flatItems) {
		end = len(m.flatItems)
	}

	for i := start; i < end; i++ {
		item := m.flatItems[i]
		selected := !m.searchFocused && i == m.cursor

		if item.isHeader {
			arrow := "▼"
			if item.collapsed {
				arrow = "▸"
			}
			label := fmt.Sprintf("%s %s (%d)", arrow, item.groupName, item.projectCount)
			if selected {
				b.WriteString(groupHeaderSelectedStyle.Render(label) + "\n")
			} else {
				b.WriteString(groupHeaderStyle.Render(label) + "\n")
			}
			continue
		}

		p := item.project
		name := p.Name
		path := shortenPath(p.Path, home)
		pType := p.ProjectType
		branch := p.GitBranch
		hidden := m.isHidden(p)

		nameCol := lipgloss.NewStyle().Width(20).Render(name)
		pathCol := pathStyle.Width(30).Render(path)
		typeCol := typeStyle.Width(8).Render(pType)
		branchCol := branchStyle.Render(branch)

		row := fmt.Sprintf("%s %s %s %s", nameCol, pathCol, typeCol, branchCol)

		if hidden {
			if selected {
				b.WriteString(hiddenSelectedItemStyle.Render("  ▸ "+row) + "\n")
			} else {
				b.WriteString(hiddenItemStyle.Render("    "+row) + "\n")
			}
		} else {
			if selected {
				b.WriteString(selectedItemStyle.Render("  ▸ "+row) + "\n")
			} else {
				b.WriteString(itemStyle.Render("    "+row) + "\n")
			}
		}
	}

	if len(m.flatItems) == 0 && len(m.projects) > 0 {
		b.WriteString(helpStyle.Render("  No results found") + "\n")
	}

	b.WriteString("\n")
	if status != "" {
		b.WriteString(statusBarStyle.Render(status) + "\n")
	}
	if m.showHidden {
		b.WriteString(helpStyle.Render("(showing hidden projects)") + "\n")
	}
	var hint string
	if m.searchFocused {
		hint = "esc: clear  ↓: back to list"
	} else {
		ideName := "IDE"
		if p, ok := m.selected(); ok {
			if name := m.resolveIDEName(p); name != "" {
				ideName = name
			}
		}
		if m.showHidden {
			hint = "enter: actions  t: terminal  o: " + ideName + "  r: refresh  h: hide hidden  ←/→: collapse  s: settings  q: quit"
		} else {
			hint = "enter: actions  t: terminal  o: " + ideName + "  r: refresh  h: hidden  ←/→: collapse  s: settings  q: quit"
		}
	}
	b.WriteString(helpStyle.Render(hint))

	return b.String()
}

func shortenPath(path, home string) string {
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

type projectSource []project.Project

func (ps projectSource) String(i int) string {
	return ps[i].Name + " " + ps[i].Path
}

func (ps projectSource) Len() int {
	return len(ps)
}

func (m *listModel) removeProject(path string) {
	var kept []project.Project
	for _, p := range m.projects {
		if p.Path != path {
			kept = append(kept, p)
		}
	}
	m.projects = kept
	m.rebuildSections()
}

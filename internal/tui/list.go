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
)

type listModel struct {
	search         textinput.Model
	projects       []project.Project
	filtered       []project.Project
	cursor         int
	width          int
	height         int
	showHidden     bool
	searchFocused  bool
	isHidden       func(project.Project) bool
	resolveIDEName func(project.Project) string
}

func newListModel(isHidden func(project.Project) bool, resolveIDEName func(project.Project) string) listModel {
	ti := textinput.New()
	ti.Placeholder = "search..."
	ti.Prompt = "> "
	ti.Focus()
	return listModel{
		search:         ti,
		searchFocused:  true,
		isHidden:       isHidden,
		resolveIDEName: resolveIDEName,
	}
}

func (m *listModel) setProjects(projects []project.Project) {
	m.projects = projects
	m.applyFilter()
}

func (m *listModel) setSize(w, h int) {
	m.width = w
	m.height = h
}

func (m listModel) selected() (project.Project, bool) {
	if len(m.filtered) == 0 {
		return project.Project{}, false
	}
	return m.filtered[m.cursor], true
}

func (m listModel) Update(msg tea.Msg) (listModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchFocused {
			switch msg.String() {
			case "down":
				m.searchFocused = false
				m.search.Blur()
				return m, nil
			case "esc":
				m.search.SetValue("")
				m.applyFilter()
				return m, nil
			case "enter":
				return m, nil
			}
			oldVal := m.search.Value()
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			if m.search.Value() != oldVal {
				m.applyFilter()
			}
			return m, cmd
		}

		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.searchFocused = true
				m.search.Focus()
				return m, textinput.Blink
			}
			return m, nil
		case "down":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil
		case "esc":
			m.searchFocused = true
			m.search.Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m *listModel) applyFilter() {
	var source []project.Project
	if m.showHidden {
		source = m.projects
	} else {
		source = make([]project.Project, 0, len(m.projects))
		for _, p := range m.projects {
			if !m.isHidden(p) {
				source = append(source, p)
			}
		}
	}

	query := strings.TrimSpace(m.search.Value())
	if query == "" {
		m.filtered = source
		m.cursor = 0
		return
	}

	matches := fuzzy.FindFrom(query, projectSource(source))
	m.filtered = make([]project.Project, len(matches))
	for i, match := range matches {
		m.filtered[i] = source[match.Index]
	}
	m.cursor = 0
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
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := start; i < end; i++ {
		p := m.filtered[i]
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

		selected := !m.searchFocused && i == m.cursor
		if hidden {
			if selected {
				b.WriteString(hiddenSelectedItemStyle.Render("▸ "+row) + "\n")
			} else {
				b.WriteString(hiddenItemStyle.Render("  "+row) + "\n")
			}
		} else {
			if selected {
				b.WriteString(selectedItemStyle.Render("▸ "+row) + "\n")
			} else {
				b.WriteString(itemStyle.Render("  "+row) + "\n")
			}
		}
	}

	if len(m.filtered) == 0 && len(m.projects) > 0 {
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
			hint = "enter: actions  t: terminal  o: " + ideName + "  r: refresh  h: hide hidden  ?: help  q: quit"
		} else {
			hint = "enter: actions  t: terminal  o: " + ideName + "  r: refresh  h: hidden  ?: help  q: quit"
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
	filtered := m.projects[:0]
	for _, p := range m.projects {
		if p.Path != path {
			filtered = append(filtered, p)
		}
	}
	m.projects = filtered
	m.applyFilter()
}

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
	search   textinput.Model
	projects []project.Project
	filtered []project.Project
	cursor   int
	width    int
	height   int
}

func newListModel() listModel {
	ti := textinput.New()
	ti.Placeholder = "ara..."
	ti.Prompt = "> "
	ti.Focus()
	return listModel{
		search: ti,
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
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil
		}
	}

	oldVal := m.search.Value()
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	if m.search.Value() != oldVal {
		m.applyFilter()
	}
	return m, cmd
}

func (m *listModel) applyFilter() {
	query := strings.TrimSpace(m.search.Value())
	if query == "" {
		m.filtered = m.projects
		m.cursor = 0
		return
	}

	matches := fuzzy.FindFrom(query, projectSource(m.projects))
	m.filtered = make([]project.Project, len(matches))
	for i, match := range matches {
		m.filtered[i] = m.projects[match.Index]
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

		nameCol := lipgloss.NewStyle().Width(20).Render(name)
		pathCol := pathStyle.Width(30).Render(path)
		typeCol := typeStyle.Width(8).Render(pType)
		branchCol := branchStyle.Render(branch)

		row := fmt.Sprintf("%s %s %s %s", nameCol, pathCol, typeCol, branchCol)

		if i == m.cursor {
			b.WriteString(selectedItemStyle.Render("▸ " + row) + "\n")
		} else {
			b.WriteString(itemStyle.Render("  " + row) + "\n")
		}
	}

	if len(m.filtered) == 0 && len(m.projects) > 0 {
		b.WriteString(helpStyle.Render("  Sonuç bulunamadı") + "\n")
	}

	b.WriteString("\n")
	if status != "" {
		b.WriteString(statusBarStyle.Render(status) + "\n")
	}
	b.WriteString(helpStyle.Render("enter: eylemler  r: yenile  ?: yardım  q: çık"))

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

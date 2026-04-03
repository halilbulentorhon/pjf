package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type outputModel struct {
	title   string
	lines   []string
	scroll  int
	running bool
	height  int
}

func newOutputModel(title string, height int) outputModel {
	return outputModel{
		title:   title,
		running: true,
		height:  height,
	}
}

func (m *outputModel) setOutput(output string) {
	m.lines = strings.Split(output, "\n")
	m.running = false
	m.scroll = m.maxScroll()
}

func (m outputModel) maxScroll() int {
	visible := m.height - 6
	if visible < 1 {
		visible = 10
	}
	max := len(m.lines) - visible
	if max < 0 {
		return 0
	}
	return max
}

func (m outputModel) Update(msg tea.Msg) (outputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.scroll > 0 {
				m.scroll--
			}
		case "down", "j":
			if m.scroll < m.maxScroll() {
				m.scroll++
			}
		}
	}
	return m, nil
}

func (m outputModel) View() string {
	return actionMenuStyle.Render(func() string {
		s := titleStyle.Render(m.title) + "\n\n"

		if m.running {
			s += helpStyle.Render("Running...") + "\n"
			s += "\n" + helpStyle.Render("esc: cancel")
			return s
		}

		visible := m.height - 6
		if visible < 1 {
			visible = 10
		}

		end := m.scroll + visible
		if end > len(m.lines) {
			end = len(m.lines)
		}

		for i := m.scroll; i < end; i++ {
			s += m.lines[i] + "\n"
		}

		if len(m.lines) == 0 {
			s += helpStyle.Render("(no output)") + "\n"
		}

		s += "\n" + helpStyle.Render("↑/↓: scroll  esc: close")
		return s
	}())
}

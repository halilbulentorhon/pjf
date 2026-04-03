package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type cmdInputResult struct {
	command string
	save    bool
}

type cmdInputModel struct {
	input    textinput.Model
	choosing bool
	command  string
	cursor   int
}

func newCmdInputModel() cmdInputModel {
	ti := textinput.New()
	ti.Placeholder = "command..."
	ti.Prompt = "$ "
	ti.Focus()
	return cmdInputModel{input: ti}
}

func (m cmdInputModel) Update(msg tea.Msg) (cmdInputModel, tea.Cmd, cmdInputResult) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.choosing {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < 1 {
					m.cursor++
				}
			case "enter":
				if m.cursor == 0 {
					return m, nil, cmdInputResult{command: m.command, save: true}
				}
				return m, nil, cmdInputResult{command: m.command, save: false}
			case "esc":
				m.choosing = false
				m.input.Focus()
				return m, textinput.Blink, cmdInputResult{}
			}
			return m, nil, cmdInputResult{}
		}

		if msg.String() == "enter" {
			val := m.input.Value()
			if val != "" {
				m.command = val
				m.choosing = true
				m.cursor = 0
				m.input.Blur()
				return m, nil, cmdInputResult{}
			}
			return m, nil, cmdInputResult{}
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd, cmdInputResult{}
}

func (m cmdInputModel) View() string {
	return actionMenuStyle.Render(func() string {
		s := titleStyle.Render("Run Custom Command") + "\n\n"

		if m.choosing {
			s += helpStyle.Render("$ "+m.command) + "\n\n"
			options := []string{"Save & Run", "Run only"}
			for i, opt := range options {
				if i == m.cursor {
					s += selectedItemStyle.Render("▸ "+opt) + "\n"
				} else {
					s += itemStyle.Render("  "+opt) + "\n"
				}
			}
			s += "\n" + helpStyle.Render("enter: select  esc: back")
			return s
		}

		s += m.input.View() + "\n\n"
		s += helpStyle.Render("enter: run  esc: back")
		return s
	}())
}

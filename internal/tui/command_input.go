package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type cmdInputResult struct {
	command string
}

type cmdInputModel struct {
	input textinput.Model
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
		if msg.String() == "enter" {
			val := m.input.Value()
			if val != "" {
				return m, nil, cmdInputResult{command: val}
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
		s += m.input.View() + "\n\n"
		s += helpStyle.Render("enter: run  esc: back")
		return s
	}())
}

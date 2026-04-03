package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type settingsInputResult struct {
	value     string
	cancelled bool
}

type settingsInputModel struct {
	input textinput.Model
	label string
}

func newSettingsInput(label, placeholder string) settingsInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = "> "
	ti.Focus()
	return settingsInputModel{input: ti, label: label}
}

func (m settingsInputModel) Update(msg tea.Msg) (settingsInputModel, tea.Cmd, settingsInputResult) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			val := m.input.Value()
			if val != "" {
				return m, nil, settingsInputResult{value: val}
			}
			return m, nil, settingsInputResult{}
		case "esc":
			return m, nil, settingsInputResult{cancelled: true}
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd, settingsInputResult{}
}

func (m settingsInputModel) View() string {
	return actionMenuStyle.Render(func() string {
		s := titleStyle.Render(m.label) + "\n\n"
		s += m.input.View() + "\n\n"
		s += helpStyle.Render("enter: confirm  esc: cancel")
		return s
	}())
}

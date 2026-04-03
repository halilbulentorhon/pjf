package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type wizardStep int

const (
	wizardWelcome wizardStep = iota
	wizardDirInput
)

type wizardModel struct {
	step  wizardStep
	input textinput.Model
	dirs  []string
}

func newWizardModel() wizardModel {
	ti := textinput.New()
	ti.Placeholder = "~/projects"
	ti.Focus()
	return wizardModel{
		step:  wizardWelcome,
		input: ti,
	}
}

func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m wizardModel) Update(msg tea.Msg) (wizardModel, tea.Cmd, bool) {
	switch m.step {
	case wizardWelcome:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter":
				m.step = wizardDirInput
				return m, textinput.Blink, false
			case "tab":
				home, _ := os.UserHomeDir()
				m.dirs = []string{home}
				return m, nil, true
			}
		}
		return m, nil, false

	case wizardDirInput:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter":
				val := strings.TrimSpace(m.input.Value())
				if val == "" {
					if len(m.dirs) > 0 {
						return m, nil, true
					}
					return m, nil, false
				}
				val = expandHome(val)
				m.dirs = append(m.dirs, val)
				m.input.SetValue("")
				return m, nil, false
			case "tab":
				home, _ := os.UserHomeDir()
				m.dirs = []string{home}
				return m, nil, true
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd, false
	}
	return m, nil, false
}

func (m wizardModel) View() string {
	switch m.step {
	case wizardWelcome:
		return wizardTextStyle.Render(
			titleStyle.Render("ProjectFinder (pjf)") + "\n\n" +
				"Yerel geliştirme projelerini bulan, listeleyen ve yöneten bir TUI aracı.\n" +
				"Git repolarını tarar ve hızlı erişim sağlar.\n\n" +
				"enter: dizin seçimine geç   tab: atla & tümünü tara",
		)
	case wizardDirInput:
		var b strings.Builder
		b.WriteString(titleStyle.Render("Hangi dizinleri taramak istiyorsun?") + "\n\n")
		for _, d := range m.dirs {
			b.WriteString("  + " + d + "\n")
		}
		b.WriteString("\n" + m.input.View() + "\n\n")
		b.WriteString(helpStyle.Render("enter: ekle (boş bırak = bitir)   tab: atla & tümünü tara"))
		return wizardTextStyle.Render(b.String())
	}
	return ""
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return home + path[1:]
	}
	if path == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	return path
}

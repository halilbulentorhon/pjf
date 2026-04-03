package tui

import (
	"os"
	"path/filepath"
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
	step        wizardStep
	input       textinput.Model
	dirs        []string
	completions []string
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
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			m.step = wizardDirInput
			return m, textinput.Blink, false
		}
		return m, nil, false

	case wizardDirInput:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.Type {
			case tea.KeyEnter:
				m.completions = nil
				val := strings.TrimSpace(m.input.Value())
				if val == "" {
					if len(m.dirs) > 0 {
						return m, nil, true
					}
					home, _ := os.UserHomeDir()
					m.dirs = []string{home}
					return m, nil, true
				}
				val = expandHome(val)
				m.dirs = append(m.dirs, val)
				m.input.SetValue("")
				return m, nil, false
			case tea.KeyTab:
				m.completions = nil
				val := m.input.Value()
				if val == "" {
					return m, nil, false
				}
				completed, matches := completeDir(val)
				if completed != val {
					m.input.SetValue(completed)
					m.input.SetCursor(len(completed))
				}
				if len(matches) > 1 {
					m.completions = matches
				}
				return m, nil, false
			case tea.KeyBackspace, tea.KeyCtrlH:
				if len(m.input.Value()) == 0 && len(m.dirs) > 0 {
					m.dirs = m.dirs[:len(m.dirs)-1]
					m.completions = nil
					return m, nil, false
				}
				m.completions = nil
			default:
				m.completions = nil
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
				"A TUI tool for discovering, listing, and managing local dev projects.\n" +
				"Scans for git repos and provides quick access.\n\n" +
				"enter: continue",
		)
	case wizardDirInput:
		var b strings.Builder
		b.WriteString(titleStyle.Render("Which directories should be scanned?") + "\n\n")
		for _, d := range m.dirs {
			b.WriteString("  + " + d + "\n")
		}
		b.WriteString("\n" + m.input.View() + "\n")
		if len(m.completions) > 0 {
			b.WriteString(helpStyle.Render("  " + strings.Join(m.completions, "  ")) + "\n")
		}
		b.WriteString("\n")
		hints := "tab: complete  enter: add (empty = done)"
		if len(m.dirs) > 0 {
			hints += "  backspace: remove last"
		}
		b.WriteString(helpStyle.Render(hints))
		return wizardTextStyle.Render(b.String())
	}
	return ""
}

func completeDir(input string) (string, []string) {
	expanded := expandHome(input)

	info, err := os.Stat(expanded)
	if err == nil && info.IsDir() && strings.HasSuffix(expanded, "/") {
		entries, err := os.ReadDir(expanded)
		if err != nil {
			return input, nil
		}
		var dirs []string
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				dirs = append(dirs, e.Name())
			}
		}
		if len(dirs) == 1 {
			result := input + dirs[0] + "/"
			return result, nil
		}
		return input, dirs
	}

	dir := filepath.Dir(expanded)
	base := filepath.Base(expanded)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return input, nil
	}

	var matches []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") && strings.HasPrefix(strings.ToLower(e.Name()), strings.ToLower(base)) {
			matches = append(matches, e.Name())
		}
	}

	if len(matches) == 0 {
		return input, nil
	}

	if len(matches) == 1 {
		prefix := input[:len(input)-len(base)]
		return prefix + matches[0] + "/", nil
	}

	common := matches[0]
	for _, m := range matches[1:] {
		common = commonPrefix(common, m)
	}
	if len(common) > len(base) {
		prefix := input[:len(input)-len(base)]
		return prefix + common, matches
	}
	return input, matches
}

func commonPrefix(a, b string) string {
	la, lb := strings.ToLower(a), strings.ToLower(b)
	i := 0
	for i < len(la) && i < len(lb) && la[i] == lb[i] {
		i++
	}
	return a[:i]
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

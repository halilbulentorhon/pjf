package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Padding(0, 1)

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			Padding(0, 2)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("12")).
				Bold(true).
				Padding(0, 2)

	hiddenItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			Padding(0, 2)

	hiddenSelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Italic(true).
				Bold(true).
				Padding(0, 2)

	confirmStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Bold(true)

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	defaultMarkerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("3"))

	activeTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true).
			Underline(true)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	typeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))

	branchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Padding(1, 1, 0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)

	wizardTextStyle = lipgloss.NewStyle().
			Padding(1, 2)

	actionMenuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)
)

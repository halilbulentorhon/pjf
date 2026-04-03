package tui

type helpModel struct{}

func newHelpModel() helpModel {
	return helpModel{}
}

func (m helpModel) View(width, height int) string {
	content := titleStyle.Render("Keyboard Shortcuts") + "\n\n" +
		"  Any char     Fuzzy search filter\n" +
		"  ↑/↓ or j/k   Navigate list\n" +
		"  enter        Action menu\n" +
		"  r            Rescan projects\n" +
		"  ?            Toggle this help\n" +
		"  q / ctrl+c   Quit\n\n" +
		helpStyle.Render("esc or ?: close")

	return actionMenuStyle.Render(content)
}

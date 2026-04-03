package tui

type helpModel struct{}

func newHelpModel() helpModel {
	return helpModel{}
}

func (m helpModel) View(width, height int) string {
	content := titleStyle.Render("Keyboard Shortcuts") + "\n\n" +
		"  ↑/↓          Navigate list / search\n" +
		"  enter        Action menu\n" +
		"  t            Open in terminal\n" +
		"  o            Open in IDE\n" +
		"  h            Toggle hidden projects\n" +
		"  r            Rescan projects\n" +
		"  ?            Toggle this help\n" +
		"  q            Quit\n\n" +
		"  Search: ↑ at top enters search, ↓/esc returns to list\n\n" +
		helpStyle.Render("esc or ?: close")

	return actionMenuStyle.Render(content)
}

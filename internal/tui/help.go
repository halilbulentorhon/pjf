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
		"  c            Run command\n" +
		"  e            Edit mode\n" +
		"  s            Settings\n" +
		"  r            Rescan projects\n" +
		"  ←/→          Collapse/expand groups\n" +
		"  ?            Toggle this help\n" +
		"  q            Quit\n\n" +
		titleStyle.Render("Edit Mode") + "\n\n" +
		"  h            Hide/unhide project\n" +
		"  m            Move to group\n" +
		"  d            Delete project\n" +
		"  v            Show/hide hidden\n" +
		"  w/s          Reorder groups\n" +
		"  e/esc        Exit edit mode\n\n" +
		"  Search: ↑ at top enters search, ↓/esc returns to list\n\n" +
		helpStyle.Render("esc or ?: close")

	return actionMenuStyle.Render(content)
}

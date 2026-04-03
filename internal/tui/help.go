package tui

type helpModel struct{}

func newHelpModel() helpModel {
	return helpModel{}
}

func (m helpModel) View(width, height int) string {
	content := titleStyle.Render("Klavye Kısayolları") + "\n\n" +
		"  Karakter     Fuzzy arama filtresi\n" +
		"  ↑/↓ veya j/k Listeyi gezin\n" +
		"  enter        Eylem menüsü\n" +
		"  r            Projeleri yeniden tara\n" +
		"  ?            Bu yardımı göster/gizle\n" +
		"  q / ctrl+c   Çıkış\n\n" +
		helpStyle.Render("esc veya ?: kapat")

	return actionMenuStyle.Render(content)
}

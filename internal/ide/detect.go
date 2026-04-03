package ide

func DetectAll() []IDE {
	pathIDEs := detectFromPATH()
	platformIDEs := detectPlatform()

	seen := make(map[string]bool)
	var merged []IDE
	for _, i := range pathIDEs {
		seen[i.Slug] = true
		merged = append(merged, i)
	}
	for _, i := range platformIDEs {
		if !seen[i.Slug] {
			merged = append(merged, i)
		}
	}
	return merged
}

func FindBySlug(ides []IDE, slug string) (IDE, bool) {
	for _, i := range ides {
		if i.Slug == slug {
			return i, true
		}
	}
	return IDE{}, false
}

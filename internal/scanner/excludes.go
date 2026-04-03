package scanner

var defaultExcludeList = []string{
	"node_modules", ".git", "vendor", ".cache", ".Trash",
	"Library", ".npm", ".cargo", ".gradle", ".m2",
	".venv", "__pycache__", "dist", "build", "target",
	".idea", ".vscode", ".DS_Store", "tmp", "temp", ".docker", "snap",
}

func DefaultExcludes() map[string]bool {
	m := make(map[string]bool, len(defaultExcludeList))
	for _, e := range defaultExcludeList {
		m[e] = true
	}
	return m
}

func MergeExcludes(extra []string) map[string]bool {
	m := DefaultExcludes()
	for _, e := range extra {
		m[e] = true
	}
	return m
}

//go:build darwin

package ide

import (
	"os"
	"path/filepath"
)

type appDef struct {
	slug     string
	name     string
	kind     string
	patterns []string
}

var macApps = []appDef{
	{slug: "code", name: "VS Code", kind: "gui", patterns: []string{
		"/Applications/Visual Studio Code.app",
	}},
	{slug: "cursor", name: "Cursor", kind: "gui", patterns: []string{
		"/Applications/Cursor.app",
	}},
	{slug: "idea", name: "IntelliJ IDEA", kind: "gui", patterns: []string{
		"/Applications/IntelliJ IDEA*.app",
	}},
	{slug: "goland", name: "GoLand", kind: "gui", patterns: []string{
		"/Applications/GoLand.app",
	}},
	{slug: "webstorm", name: "WebStorm", kind: "gui", patterns: []string{
		"/Applications/WebStorm.app",
	}},
	{slug: "zed", name: "Zed", kind: "gui", patterns: []string{
		"/Applications/Zed.app",
	}},
}

func detectPlatform() []IDE {
	var found []IDE
	home, _ := os.UserHomeDir()
	toolboxBase := filepath.Join(home, "Library", "Application Support", "JetBrains", "Toolbox", "apps")

	for _, app := range macApps {
		for _, pattern := range app.patterns {
			matches, _ := filepath.Glob(pattern)
			if len(matches) > 0 {
				found = append(found, IDE{
					Name: app.name,
					Slug: app.slug,
					Path: matches[0],
					Kind: app.kind,
				})
				break
			}
		}
	}

	toolboxApps := []struct {
		dir  string
		slug string
		name string
	}{
		{"IDEA-U", "idea", "IntelliJ IDEA"},
		{"IDEA-C", "idea", "IntelliJ IDEA"},
		{"GoLand", "goland", "GoLand"},
		{"WebStorm", "webstorm", "WebStorm"},
	}

	for _, ta := range toolboxApps {
		base := filepath.Join(toolboxBase, ta.dir)
		if _, err := os.Stat(base); err != nil {
			continue
		}
		entries, _ := os.ReadDir(base)
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			appPattern := filepath.Join(base, e.Name(), ta.name+"*.app")
			matches, _ := filepath.Glob(appPattern)
			if len(matches) > 0 {
				found = append(found, IDE{
					Name: ta.name,
					Slug: ta.slug,
					Path: matches[0],
					Kind: "gui",
				})
				break
			}
		}
	}

	return found
}

package ide

import "os/exec"

type IDE struct {
	Name string
	Slug string
	Path string
	Kind string
}

type ideDef struct {
	name   string
	slug   string
	binary string
	kind   string
}

var registry = []ideDef{
	{name: "VS Code", slug: "code", binary: "code", kind: "gui"},
	{name: "Cursor", slug: "cursor", binary: "cursor", kind: "gui"},
	{name: "IntelliJ IDEA", slug: "idea", binary: "idea", kind: "gui"},
	{name: "GoLand", slug: "goland", binary: "goland", kind: "gui"},
	{name: "WebStorm", slug: "webstorm", binary: "webstorm", kind: "gui"},
	{name: "Zed", slug: "zed", binary: "zed", kind: "gui"},
	{name: "Claude Code", slug: "claude", binary: "claude", kind: "terminal"},
	{name: "Neovim", slug: "nvim", binary: "nvim", kind: "terminal"},
	{name: "Vim", slug: "vim", binary: "vim", kind: "terminal"},
}

func detectFromPATH() []IDE {
	var found []IDE
	for _, def := range registry {
		path, err := exec.LookPath(def.binary)
		if err == nil {
			found = append(found, IDE{
				Name: def.name,
				Slug: def.slug,
				Path: path,
				Kind: def.kind,
			})
		}
	}
	return found
}

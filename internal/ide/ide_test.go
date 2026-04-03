package ide

import (
	"testing"
)

func TestDetectAllReturnsNoDuplicates(t *testing.T) {
	ides := DetectAll()
	seen := make(map[string]bool)
	for _, i := range ides {
		if seen[i.Slug] {
			t.Errorf("duplicate IDE slug: %s", i.Slug)
		}
		seen[i.Slug] = true
	}
}

func TestFindBySlug(t *testing.T) {
	ides := []IDE{
		{Name: "VS Code", Slug: "code"},
		{Name: "GoLand", Slug: "goland"},
	}

	found, ok := FindBySlug(ides, "goland")
	if !ok {
		t.Fatal("expected to find goland")
	}
	if found.Name != "GoLand" {
		t.Errorf("expected GoLand, got %s", found.Name)
	}

	_, ok = FindBySlug(ides, "nonexistent")
	if ok {
		t.Error("expected not to find nonexistent")
	}
}

func TestRegistryHasExpectedIDEs(t *testing.T) {
	expected := []string{"code", "cursor", "idea", "goland", "webstorm", "zed", "claude", "nvim", "vim"}
	slugs := make(map[string]bool)
	for _, def := range registry {
		slugs[def.slug] = true
	}
	for _, slug := range expected {
		if !slugs[slug] {
			t.Errorf("expected IDE %q in registry", slug)
		}
	}
}

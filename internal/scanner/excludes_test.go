package scanner

import "testing"

func TestDefaultExcludes(t *testing.T) {
	excludes := DefaultExcludes()
	expected := []string{"node_modules", ".git", "vendor", ".cache", ".venv", "__pycache__"}
	for _, e := range expected {
		if !excludes[e] {
			t.Errorf("expected %q in default excludes", e)
		}
	}
}

func TestMergeExcludes(t *testing.T) {
	merged := MergeExcludes([]string{"custom_dir", "another"})
	if !merged["custom_dir"] {
		t.Error("expected custom_dir in merged excludes")
	}
	if !merged["node_modules"] {
		t.Error("expected node_modules (default) in merged excludes")
	}
}

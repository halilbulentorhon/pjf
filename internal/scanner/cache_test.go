package scanner

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/halilbulentorhon/pjf/internal/project"
)

func TestCacheSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")
	store := &JSONCacheStore{Path: path}

	projects := []project.Project{
		{Path: "/tmp/foo", Name: "foo", GitBranch: "main", ProjectType: "go"},
	}

	err := store.Save(projects)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cache.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(cache.Projects))
	}
	if cache.Projects[0].Name != "foo" {
		t.Errorf("expected name 'foo', got %q", cache.Projects[0].Name)
	}
	if cache.LastScan.IsZero() {
		t.Error("expected LastScan to be set")
	}
}

func TestCacheLoadMissing(t *testing.T) {
	store := &JSONCacheStore{Path: "/nonexistent/cache.json"}
	cache, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cache != nil {
		t.Error("expected nil cache for missing file")
	}
}

func TestCacheIsStale(t *testing.T) {
	c := &Cache{
		LastScan: time.Now().Add(-25 * time.Hour),
	}
	if !c.IsStale(24) {
		t.Error("expected stale cache")
	}

	c.LastScan = time.Now().Add(-1 * time.Hour)
	if c.IsStale(24) {
		t.Error("expected fresh cache")
	}
}

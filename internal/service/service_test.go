package service

import (
	"context"
	"testing"
	"time"

	"github.com/halilbulentorhon/pjf/internal/config"
	"github.com/halilbulentorhon/pjf/internal/project"
	"github.com/halilbulentorhon/pjf/internal/scanner"
)

type mockScanner struct {
	projects []project.Project
}

func (m *mockScanner) Scan(_ context.Context, _ []string, _ int, _ map[string]bool, onProgress func(int)) ([]project.Project, error) {
	for i, p := range m.projects {
		_ = p
		if onProgress != nil {
			onProgress(i + 1)
		}
	}
	return m.projects, nil
}

type mockCache struct {
	cache *scanner.Cache
	saved []project.Project
}

func (m *mockCache) Load() (*scanner.Cache, error) {
	return m.cache, nil
}

func (m *mockCache) Save(projects []project.Project) error {
	m.saved = projects
	return nil
}

func TestLoadOrScanNoCache(t *testing.T) {
	svc := New(
		&config.Config{MaxDepth: 5, CacheTTL: 24},
		&mockScanner{},
		&mockCache{cache: nil},
	)
	projects, fromCache, needsRefresh, err := svc.LoadOrScan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if fromCache {
		t.Error("expected fromCache=false")
	}
	if needsRefresh {
		t.Error("expected needsRefresh=false")
	}
	if projects != nil {
		t.Error("expected nil projects")
	}
}

func TestLoadOrScanFreshCache(t *testing.T) {
	cache := &scanner.Cache{
		Projects: []project.Project{{Name: "test"}},
		LastScan: time.Now(),
	}

	svc := New(
		&config.Config{MaxDepth: 5, CacheTTL: 24},
		&mockScanner{},
		&mockCache{cache: cache},
	)
	projects, fromCache, needsRefresh, err := svc.LoadOrScan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !fromCache {
		t.Error("expected fromCache=true")
	}
	if needsRefresh {
		t.Error("expected needsRefresh=false for fresh cache")
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
}

func TestScanSavesToCache(t *testing.T) {
	mc := &mockCache{cache: nil}
	ms := &mockScanner{projects: []project.Project{{Name: "found"}}}
	svc := New(
		&config.Config{ScanDirs: []string{"/tmp"}, MaxDepth: 5, CacheTTL: 24},
		ms,
		mc,
	)
	projects, err := svc.Scan(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
	if len(mc.saved) != 1 {
		t.Error("expected cache to be saved")
	}
}

func TestHideProject(t *testing.T) {
	cfg := &config.Config{MaxDepth: 5, CacheTTL: 24}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	p := project.Project{Path: "/tmp/test-project", Name: "test"}

	svc.HideProject(p)

	if !svc.IsHidden(p) {
		t.Error("expected project to be hidden")
	}
	if len(svc.Cfg.HiddenProjects) != 1 {
		t.Errorf("expected 1 hidden project, got %d", len(svc.Cfg.HiddenProjects))
	}
}

func TestHideProjectIdempotent(t *testing.T) {
	cfg := &config.Config{MaxDepth: 5, CacheTTL: 24}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	p := project.Project{Path: "/tmp/test-project", Name: "test"}

	svc.HideProject(p)
	svc.HideProject(p)

	if len(svc.Cfg.HiddenProjects) != 1 {
		t.Errorf("expected 1 hidden project after double hide, got %d", len(svc.Cfg.HiddenProjects))
	}
}

func TestUnhideProject(t *testing.T) {
	cfg := &config.Config{
		MaxDepth:       5,
		CacheTTL:       24,
		HiddenProjects: []string{"/tmp/test-project"},
	}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	p := project.Project{Path: "/tmp/test-project", Name: "test"}

	svc.UnhideProject(p)

	if svc.IsHidden(p) {
		t.Error("expected project to not be hidden after unhide")
	}
	if len(svc.Cfg.HiddenProjects) != 0 {
		t.Errorf("expected 0 hidden projects, got %d", len(svc.Cfg.HiddenProjects))
	}
}

func TestIsHidden(t *testing.T) {
	cfg := &config.Config{
		MaxDepth:       5,
		CacheTTL:       24,
		HiddenProjects: []string{"/tmp/hidden-one"},
	}
	svc := New(cfg, &mockScanner{}, &mockCache{})

	if !svc.IsHidden(project.Project{Path: "/tmp/hidden-one"}) {
		t.Error("expected hidden-one to be hidden")
	}
	if svc.IsHidden(project.Project{Path: "/tmp/visible-one"}) {
		t.Error("expected visible-one to not be hidden")
	}
}

func TestDeleteProjectCleansHidden(t *testing.T) {
	cfg := &config.Config{
		MaxDepth:       5,
		CacheTTL:       24,
		HiddenProjects: []string{"/tmp/nonexistent-for-delete-test"},
	}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	p := project.Project{Path: "/tmp/nonexistent-for-delete-test", Name: "test"}

	_ = svc.DeleteProject(p)

	if svc.IsHidden(p) {
		t.Error("expected hidden entry to be cleaned after delete")
	}
}

func TestRemoveFromCache(t *testing.T) {
	mc := &mockCache{
		cache: &scanner.Cache{
			Projects: []project.Project{
				{Path: "/tmp/a", Name: "a"},
				{Path: "/tmp/b", Name: "b"},
			},
		},
	}
	cfg := &config.Config{MaxDepth: 5, CacheTTL: 24}
	svc := New(cfg, &mockScanner{}, mc)

	err := svc.RemoveFromCache(project.Project{Path: "/tmp/a"})
	if err != nil {
		t.Fatal(err)
	}
	if len(mc.saved) != 1 {
		t.Errorf("expected 1 project in cache, got %d", len(mc.saved))
	}
	if mc.saved[0].Path != "/tmp/b" {
		t.Errorf("expected /tmp/b to remain, got %s", mc.saved[0].Path)
	}
}

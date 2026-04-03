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

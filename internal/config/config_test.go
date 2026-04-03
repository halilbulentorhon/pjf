package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFirstRun(t *testing.T) {
	dir := t.TempDir()
	cfg, isFirstRun, err := Load(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !isFirstRun {
		t.Error("expected isFirstRun=true for missing config")
	}
	if cfg.MaxDepth != 5 {
		t.Errorf("expected default MaxDepth=5, got %d", cfg.MaxDepth)
	}
	if cfg.CacheTTL != 24 {
		t.Errorf("expected default CacheTTL=24, got %d", cfg.CacheTTL)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pjf", "config.yaml")

	cfg := &Config{
		ScanDirs: []string{"/home/user/projects", "/opt/work"},
		MaxDepth: 3,
		CacheTTL: 12,
	}
	err := Save(path, cfg)
	if err != nil {
		t.Fatal(err)
	}

	loaded, isFirstRun, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if isFirstRun {
		t.Error("expected isFirstRun=false for existing config")
	}
	if len(loaded.ScanDirs) != 2 {
		t.Errorf("expected 2 scan dirs, got %d", len(loaded.ScanDirs))
	}
	if loaded.MaxDepth != 3 {
		t.Errorf("expected MaxDepth=3, got %d", loaded.MaxDepth)
	}
}

func TestLoadAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("scan_dirs:\n  - /tmp\n"), 0644)

	cfg, _, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxDepth != 5 {
		t.Errorf("expected default MaxDepth=5, got %d", cfg.MaxDepth)
	}
	if cfg.CacheTTL != 24 {
		t.Errorf("expected default CacheTTL=24, got %d", cfg.CacheTTL)
	}
}

package service

import (
	"context"
	"testing"
	"time"

	"github.com/halilbulentorhon/pjf/internal/config"
	"github.com/halilbulentorhon/pjf/internal/ide"
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

func TestResolveIDEProjectOverride(t *testing.T) {
	cfg := &config.Config{
		MaxDepth:    5,
		CacheTTL:    24,
		DefaultIDEs: map[string]string{"go": "goland", "_default": "code"},
		ProjectIDEs: map[string]string{"/tmp/special": "zed"},
	}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	svc.SetDetectedIDEs([]ide.IDE{
		{Name: "GoLand", Slug: "goland"},
		{Name: "VS Code", Slug: "code"},
		{Name: "Zed", Slug: "zed"},
	})

	p := project.Project{Path: "/tmp/special", ProjectType: "go"}
	resolved, ok := svc.ResolveIDE(p)
	if !ok {
		t.Fatal("expected to resolve IDE")
	}
	if resolved.Slug != "zed" {
		t.Errorf("expected zed (project override), got %s", resolved.Slug)
	}
}

func TestResolveIDETypeDefault(t *testing.T) {
	cfg := &config.Config{
		MaxDepth:    5,
		CacheTTL:    24,
		DefaultIDEs: map[string]string{"go": "goland", "_default": "code"},
	}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	svc.SetDetectedIDEs([]ide.IDE{
		{Name: "GoLand", Slug: "goland"},
		{Name: "VS Code", Slug: "code"},
	})

	p := project.Project{Path: "/tmp/myapp", ProjectType: "go"}
	resolved, ok := svc.ResolveIDE(p)
	if !ok {
		t.Fatal("expected to resolve IDE")
	}
	if resolved.Slug != "goland" {
		t.Errorf("expected goland (type default), got %s", resolved.Slug)
	}
}

func TestResolveIDEFallback(t *testing.T) {
	cfg := &config.Config{
		MaxDepth:    5,
		CacheTTL:    24,
		DefaultIDEs: map[string]string{"_default": "code"},
	}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	svc.SetDetectedIDEs([]ide.IDE{
		{Name: "VS Code", Slug: "code"},
	})

	p := project.Project{Path: "/tmp/unknown", ProjectType: "unknown"}
	resolved, ok := svc.ResolveIDE(p)
	if !ok {
		t.Fatal("expected to resolve IDE")
	}
	if resolved.Slug != "code" {
		t.Errorf("expected code (_default), got %s", resolved.Slug)
	}
}

func TestResolveIDENoIDEs(t *testing.T) {
	cfg := &config.Config{MaxDepth: 5, CacheTTL: 24}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	svc.SetDetectedIDEs(nil)

	p := project.Project{Path: "/tmp/x"}
	_, ok := svc.ResolveIDE(p)
	if ok {
		t.Error("expected no IDE resolved")
	}
}

func TestSetProjectIDE(t *testing.T) {
	cfg := &config.Config{MaxDepth: 5, CacheTTL: 24}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	p := project.Project{Path: "/tmp/myapp"}

	svc.SetProjectIDE(p, "goland")

	if svc.Cfg.ProjectIDEs["/tmp/myapp"] != "goland" {
		t.Error("expected project IDE to be set")
	}
}

func TestResolveCommandsProjectAndGlobal(t *testing.T) {
	cfg := &config.Config{
		MaxDepth: 5,
		CacheTTL: 24,
		GlobalCommands: []config.CommandDef{
			{Name: "Git Status", Command: "git status"},
			{Name: "Git Pull", Command: "git pull"},
		},
		ProjectCommands: []config.ProjectCommandSet{
			{
				Path: "/tmp/myapp",
				Commands: []config.CommandDef{
					{Name: "Run Tests", Command: "go test ./..."},
					{Name: "Git Status", Command: "git status -s"},
				},
			},
		},
	}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	p := project.Project{Path: "/tmp/myapp"}

	cmds := svc.ResolveCommands(p)
	if len(cmds) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(cmds))
	}
	if cmds[0].Name != "Run Tests" || !cmds[0].IsProject {
		t.Error("expected first command to be project Run Tests")
	}
	if cmds[1].Name != "Git Status" || !cmds[1].IsProject {
		t.Error("expected second command to be project Git Status (overrides global)")
	}
	if cmds[2].Name != "Git Pull" || cmds[2].IsProject {
		t.Error("expected third command to be global Git Pull")
	}
}

func TestResolveCommandsNoProject(t *testing.T) {
	cfg := &config.Config{
		MaxDepth: 5,
		CacheTTL: 24,
		GlobalCommands: []config.CommandDef{
			{Name: "Git Status", Command: "git status"},
		},
	}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	p := project.Project{Path: "/tmp/other"}

	cmds := svc.ResolveCommands(p)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].IsProject {
		t.Error("expected global command")
	}
}

func TestRunCommand(t *testing.T) {
	cfg := &config.Config{MaxDepth: 5, CacheTTL: 24}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	p := project.Project{Path: "/tmp"}

	output, err := svc.RunCommand(p, "echo hello")
	if err != nil {
		t.Fatal(err)
	}
	if output != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", output)
	}
}

func TestRemoveProjectIDE(t *testing.T) {
	cfg := &config.Config{
		MaxDepth:    5,
		CacheTTL:    24,
		ProjectIDEs: map[string]string{"/tmp/a": "code", "/tmp/b": "goland"},
	}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	svc.RemoveProjectIDE(project.Project{Path: "/tmp/a"})
	if _, ok := svc.Cfg.ProjectIDEs["/tmp/a"]; ok {
		t.Error("expected /tmp/a to be removed")
	}
	if svc.Cfg.ProjectIDEs["/tmp/b"] != "goland" {
		t.Error("expected /tmp/b to remain")
	}
}

func TestRemoveSavedCommand(t *testing.T) {
	cfg := &config.Config{
		MaxDepth: 5,
		CacheTTL: 24,
		ProjectCommands: []config.ProjectCommandSet{
			{Path: "/tmp/a", Commands: []config.CommandDef{
				{Name: "test", Command: "go test"},
				{Name: "build", Command: "go build"},
			}},
		},
	}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	svc.RemoveSavedCommand(project.Project{Path: "/tmp/a"}, 0)
	cmds := svc.Cfg.ProjectCommands[0].Commands
	if len(cmds) != 1 || cmds[0].Name != "build" {
		t.Errorf("expected only build remaining, got %v", cmds)
	}
}

func TestAddRemoveGlobalCommand(t *testing.T) {
	cfg := &config.Config{MaxDepth: 5, CacheTTL: 24}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	svc.AddGlobalCommand("Git Status", "git status")
	svc.AddGlobalCommand("Git Pull", "git pull")
	if len(svc.Cfg.GlobalCommands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(svc.Cfg.GlobalCommands))
	}
	svc.RemoveGlobalCommand(0)
	if len(svc.Cfg.GlobalCommands) != 1 || svc.Cfg.GlobalCommands[0].Name != "Git Pull" {
		t.Error("expected Git Pull remaining")
	}
}

func TestAddRemoveScanDir(t *testing.T) {
	cfg := &config.Config{MaxDepth: 5, CacheTTL: 24, ScanDirs: []string{"/a"}}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	svc.AddScanDir("/b")
	if len(svc.Cfg.ScanDirs) != 2 {
		t.Fatal("expected 2 scan dirs")
	}
	svc.RemoveScanDir(0)
	if len(svc.Cfg.ScanDirs) != 1 || svc.Cfg.ScanDirs[0] != "/b" {
		t.Error("expected /b remaining")
	}
}

func TestAddRemoveExclude(t *testing.T) {
	cfg := &config.Config{MaxDepth: 5, CacheTTL: 24}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	svc.AddExclude("/tmp/skip")
	if len(svc.Cfg.ExtraExcludes) != 1 {
		t.Fatal("expected 1 exclude")
	}
	svc.RemoveExclude(0)
	if len(svc.Cfg.ExtraExcludes) != 0 {
		t.Error("expected empty excludes")
	}
}

func TestSetMaxDepthAndCacheTTL(t *testing.T) {
	cfg := &config.Config{MaxDepth: 5, CacheTTL: 24}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	svc.SetMaxDepth(10)
	svc.SetCacheTTL(48)
	if svc.Cfg.MaxDepth != 10 {
		t.Errorf("expected 10, got %d", svc.Cfg.MaxDepth)
	}
	if svc.Cfg.CacheTTL != 48 {
		t.Errorf("expected 48, got %d", svc.Cfg.CacheTTL)
	}
}

func TestSetRemoveDefaultIDE(t *testing.T) {
	cfg := &config.Config{MaxDepth: 5, CacheTTL: 24}
	svc := New(cfg, &mockScanner{}, &mockCache{})
	svc.SetDefaultIDE("go", "goland")
	svc.SetDefaultIDE("_default", "code")
	if svc.Cfg.DefaultIDEs["go"] != "goland" {
		t.Error("expected go -> goland")
	}
	svc.RemoveDefaultIDE("go")
	if _, ok := svc.Cfg.DefaultIDEs["go"]; ok {
		t.Error("expected go mapping removed")
	}
	if svc.Cfg.DefaultIDEs["_default"] != "code" {
		t.Error("expected _default to remain")
	}
}

package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func createGitRepo(t *testing.T, base, name string) string {
	t.Helper()
	dir := filepath.Join(base, name)
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0644)
	return dir
}

func TestScanFindsGitRepos(t *testing.T) {
	base := t.TempDir()
	createGitRepo(t, base, "project-a")
	createGitRepo(t, base, "project-b")
	os.MkdirAll(filepath.Join(base, "not-a-project"), 0755)

	s := &FileScanner{}
	projects, err := s.Scan(context.Background(), []string{base}, 5, DefaultExcludes(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
}

func TestScanReadsGitBranch(t *testing.T) {
	base := t.TempDir()
	dir := createGitRepo(t, base, "myrepo")
	os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref: refs/heads/develop\n"), 0644)

	s := &FileScanner{}
	projects, _ := s.Scan(context.Background(), []string{base}, 5, DefaultExcludes(), nil)
	if len(projects) != 1 {
		t.Fatal("expected 1 project")
	}
	if projects[0].GitBranch != "develop" {
		t.Errorf("expected branch 'develop', got %q", projects[0].GitBranch)
	}
}

func TestScanRespectsExcludes(t *testing.T) {
	base := t.TempDir()
	createGitRepo(t, base, "good-project")
	excluded := filepath.Join(base, "node_modules", "some-pkg")
	os.MkdirAll(filepath.Join(excluded, ".git"), 0755)

	s := &FileScanner{}
	projects, _ := s.Scan(context.Background(), []string{base}, 5, DefaultExcludes(), nil)
	if len(projects) != 1 {
		t.Fatalf("expected 1 project (excluded node_modules), got %d", len(projects))
	}
}

func TestScanRespectsMaxDepth(t *testing.T) {
	base := t.TempDir()
	deep := filepath.Join(base, "a", "b", "c", "d", "e", "repo")
	os.MkdirAll(filepath.Join(deep, ".git"), 0755)

	s := &FileScanner{}
	projects, _ := s.Scan(context.Background(), []string{base}, 3, DefaultExcludes(), nil)
	if len(projects) != 0 {
		t.Fatalf("expected 0 projects at depth>3, got %d", len(projects))
	}
}

func TestScanSkipsNestedRepos(t *testing.T) {
	base := t.TempDir()
	parent := createGitRepo(t, base, "parent")
	os.MkdirAll(filepath.Join(parent, "child", ".git"), 0755)

	s := &FileScanner{}
	projects, _ := s.Scan(context.Background(), []string{base}, 5, DefaultExcludes(), nil)
	if len(projects) != 1 {
		t.Fatalf("expected 1 project (parent only), got %d", len(projects))
	}
}

func TestScanCallsProgressCallback(t *testing.T) {
	base := t.TempDir()
	createGitRepo(t, base, "p1")
	createGitRepo(t, base, "p2")

	var calls []int
	s := &FileScanner{}
	s.Scan(context.Background(), []string{base}, 5, DefaultExcludes(), func(found int) {
		calls = append(calls, found)
	})
	if len(calls) < 2 {
		t.Errorf("expected at least 2 progress calls, got %d", len(calls))
	}
}

func TestScanDetectsProjectType(t *testing.T) {
	base := t.TempDir()
	dir := createGitRepo(t, base, "go-app")
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example"), 0644)

	s := &FileScanner{}
	projects, _ := s.Scan(context.Background(), []string{base}, 5, DefaultExcludes(), nil)
	if projects[0].ProjectType != "go" {
		t.Errorf("expected type 'go', got %q", projects[0].ProjectType)
	}
}

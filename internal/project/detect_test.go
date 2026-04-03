package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectType(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{"go project", []string{"go.mod"}, "go"},
		{"node project", []string{"package.json"}, "node"},
		{"java maven", []string{"pom.xml"}, "java"},
		{"java gradle", []string{"build.gradle"}, "java"},
		{"rust project", []string{"Cargo.toml"}, "rust"},
		{"python project", []string{"pyproject.toml"}, "python"},
		{"docker compose yml", []string{"docker-compose.yml"}, "docker"},
		{"docker compose yaml", []string{"docker-compose.yaml"}, "docker"},
		{"compose yml", []string{"compose.yml"}, "docker"},
		{"compose yaml", []string{"compose.yaml"}, "docker"},
		{"unknown with makefile", []string{"Makefile"}, "unknown"},
		{"empty dir", []string{}, "unknown"},
		{"go takes priority over node", []string{"go.mod", "package.json"}, "go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tt.files {
				os.WriteFile(filepath.Join(dir, f), []byte{}, 0644)
			}
			got := DetectType(dir)
			if got != tt.expected {
				t.Errorf("DetectType() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestHasDockerCompose(t *testing.T) {
	dir := t.TempDir()
	if HasDockerCompose(dir) {
		t.Error("expected false for empty dir")
	}
	os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte{}, 0644)
	if !HasDockerCompose(dir) {
		t.Error("expected true with docker-compose.yml")
	}
}

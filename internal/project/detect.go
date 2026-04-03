package project

import (
	"os"
	"path/filepath"
)

var composeFiles = []string{
	"docker-compose.yml",
	"docker-compose.yaml",
	"compose.yml",
	"compose.yaml",
}

var typeMarkers = []struct {
	file     string
	projType string
}{
	{"go.mod", "go"},
	{"package.json", "node"},
	{"pom.xml", "java"},
	{"build.gradle", "java"},
	{"Cargo.toml", "rust"},
	{"pyproject.toml", "python"},
}

func DetectType(dir string) string {
	for _, m := range typeMarkers {
		if fileExists(filepath.Join(dir, m.file)) {
			return m.projType
		}
	}
	if HasDockerCompose(dir) {
		return "docker"
	}
	return "unknown"
}

func HasDockerCompose(dir string) bool {
	for _, f := range composeFiles {
		if fileExists(filepath.Join(dir, f)) {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

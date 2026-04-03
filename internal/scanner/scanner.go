package scanner

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/halilbulentorhon/pjf/internal/project"
)

type FileScanner struct{}

func (s *FileScanner) Scan(
	ctx context.Context,
	dirs []string,
	maxDepth int,
	excludes map[string]bool,
	onProgress func(int),
) ([]project.Project, error) {
	var mu sync.Mutex
	var results []project.Project
	var wg sync.WaitGroup

	for _, root := range dirs {
		root := root
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.walkDir(ctx, root, root, 0, maxDepth, excludes, func(p project.Project) {
				mu.Lock()
				results = append(results, p)
				count := len(results)
				mu.Unlock()
				if onProgress != nil {
					onProgress(count)
				}
			})
		}()
	}

	wg.Wait()
	return results, ctx.Err()
}

func (s *FileScanner) walkDir(
	ctx context.Context,
	root, dir string,
	depth, maxDepth int,
	excludes map[string]bool,
	onFound func(project.Project),
) {
	if ctx.Err() != nil {
		return
	}
	if depth > maxDepth {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	hasGit := false
	for _, e := range entries {
		if e.Name() == ".git" && e.IsDir() {
			hasGit = true
			break
		}
	}

	if hasGit {
		p := s.buildProject(dir)
		onFound(p)
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if e.Type()&fs.ModeSymlink != 0 {
			continue
		}
		name := e.Name()
		if excludes[name] {
			continue
		}
		s.walkDir(ctx, root, filepath.Join(dir, name), depth+1, maxDepth, excludes, onFound)
	}
}

func (s *FileScanner) buildProject(dir string) project.Project {
	name := filepath.Base(dir)
	branch := readGitBranch(filepath.Join(dir, ".git", "HEAD"))
	projType := project.DetectType(dir)
	hasDC := project.HasDockerCompose(dir)

	info, _ := os.Stat(dir)
	lastMod := info.ModTime()

	return project.Project{
		Path:             dir,
		Name:             name,
		GitBranch:        branch,
		HasDockerCompose: hasDC,
		ProjectType:      projType,
		LastModified:     lastMod,
	}
}

func readGitBranch(headPath string) string {
	data, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}
	content := strings.TrimSpace(string(data))
	if strings.HasPrefix(content, "ref: refs/heads/") {
		return strings.TrimPrefix(content, "ref: refs/heads/")
	}
	if len(content) >= 7 {
		return content[:7]
	}
	return ""
}

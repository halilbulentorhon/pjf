package service

import (
	"context"

	"github.com/halilbulentorhon/pjf/internal/action"
	"github.com/halilbulentorhon/pjf/internal/config"
	"github.com/halilbulentorhon/pjf/internal/project"
	"github.com/halilbulentorhon/pjf/internal/scanner"
)

type Scanner interface {
	Scan(ctx context.Context, dirs []string, maxDepth int, excludes map[string]bool, onProgress func(int)) ([]project.Project, error)
}

type CacheStore interface {
	Load() (*scanner.Cache, error)
	Save(projects []project.Project) error
}

type ProjectService struct {
	Cfg      *config.Config
	scanner  Scanner
	cache    CacheStore
	terminal action.TerminalOpener
	clip     action.Clipboard
}

func New(cfg *config.Config, s Scanner, c CacheStore) *ProjectService {
	return &ProjectService{
		Cfg:      cfg,
		scanner:  s,
		cache:    c,
		terminal: action.NewTerminalOpener(),
		clip:     action.NewClipboard(),
	}
}

func (s *ProjectService) LoadOrScan(ctx context.Context) ([]project.Project, bool, bool, error) {
	c, err := s.cache.Load()
	if err != nil {
		return nil, false, false, err
	}
	if c == nil {
		return nil, false, false, nil
	}
	needsRefresh := c.IsStale(s.Cfg.CacheTTL)
	return c.Projects, true, needsRefresh, nil
}

func (s *ProjectService) Scan(ctx context.Context, onProgress func(int)) ([]project.Project, error) {
	excludes := scanner.MergeExcludes(s.Cfg.ExtraExcludes)
	projects, err := s.scanner.Scan(ctx, s.Cfg.ScanDirs, s.Cfg.MaxDepth, excludes, onProgress)
	if err != nil {
		return nil, err
	}
	if err := s.cache.Save(projects); err != nil {
		return projects, err
	}
	return projects, nil
}

func (s *ProjectService) Refresh(ctx context.Context) ([]project.Project, error) {
	return s.Scan(ctx, nil)
}

func (s *ProjectService) OpenTerminal(p project.Project) error {
	return s.terminal.Open(p.Path)
}

func (s *ProjectService) CopyPath(p project.Project) error {
	return s.clip.Copy(p.Path)
}

func (s *ProjectService) SaveConfig(path string) error {
	return config.Save(path, s.Cfg)
}

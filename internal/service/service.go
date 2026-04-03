package service

import (
	"context"
	"os"
	"os/exec"

	"github.com/halilbulentorhon/pjf/internal/action"
	"github.com/halilbulentorhon/pjf/internal/config"
	"github.com/halilbulentorhon/pjf/internal/ide"
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
	Cfg          *config.Config
	scanner      Scanner
	cache        CacheStore
	terminal     action.TerminalOpener
	clip         action.Clipboard
	detectedIDEs []ide.IDE
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

func (s *ProjectService) HideProject(p project.Project) {
	for _, h := range s.Cfg.HiddenProjects {
		if h == p.Path {
			return
		}
	}
	s.Cfg.HiddenProjects = append(s.Cfg.HiddenProjects, p.Path)
}

func (s *ProjectService) UnhideProject(p project.Project) {
	filtered := s.Cfg.HiddenProjects[:0]
	for _, h := range s.Cfg.HiddenProjects {
		if h != p.Path {
			filtered = append(filtered, h)
		}
	}
	s.Cfg.HiddenProjects = filtered
}

func (s *ProjectService) DeleteProject(p project.Project) error {
	if err := os.RemoveAll(p.Path); err != nil {
		return err
	}
	s.UnhideProject(p)
	return nil
}

func (s *ProjectService) IsHidden(p project.Project) bool {
	for _, h := range s.Cfg.HiddenProjects {
		if h == p.Path {
			return true
		}
	}
	return false
}

func (s *ProjectService) RemoveFromCache(p project.Project) error {
	c, err := s.cache.Load()
	if err != nil || c == nil {
		return err
	}
	filtered := c.Projects[:0]
	for _, proj := range c.Projects {
		if proj.Path != p.Path {
			filtered = append(filtered, proj)
		}
	}
	return s.cache.Save(filtered)
}

func (s *ProjectService) SetDetectedIDEs(ides []ide.IDE) {
	s.detectedIDEs = ides
}

func (s *ProjectService) DetectedIDEs() []ide.IDE {
	return s.detectedIDEs
}

func (s *ProjectService) ResolveIDE(p project.Project) (ide.IDE, bool) {
	if len(s.detectedIDEs) == 0 {
		return ide.IDE{}, false
	}

	if s.Cfg.ProjectIDEs != nil {
		if slug, ok := s.Cfg.ProjectIDEs[p.Path]; ok {
			if found, ok := ide.FindBySlug(s.detectedIDEs, slug); ok {
				return found, true
			}
		}
	}

	if s.Cfg.DefaultIDEs != nil {
		if slug, ok := s.Cfg.DefaultIDEs[p.ProjectType]; ok {
			if found, ok := ide.FindBySlug(s.detectedIDEs, slug); ok {
				return found, true
			}
		}
		if slug, ok := s.Cfg.DefaultIDEs["_default"]; ok {
			if found, ok := ide.FindBySlug(s.detectedIDEs, slug); ok {
				return found, true
			}
		}
	}

	return s.detectedIDEs[0], true
}

func (s *ProjectService) OpenIDE(p project.Project, i ide.IDE) error {
	switch i.Kind {
	case "terminal":
		return s.terminal.OpenWithCommand(p.Path, i.Slug)
	default:
		if _, err := exec.LookPath(i.Slug); err == nil {
			return exec.Command(i.Slug, p.Path).Start()
		}
		if i.Path != "" {
			return exec.Command("open", "-a", i.Path, p.Path).Start()
		}
		return exec.Command(i.Slug, p.Path).Start()
	}
}

func (s *ProjectService) SetProjectIDE(p project.Project, ideSlug string) {
	if s.Cfg.ProjectIDEs == nil {
		s.Cfg.ProjectIDEs = make(map[string]string)
	}
	s.Cfg.ProjectIDEs[p.Path] = ideSlug
}

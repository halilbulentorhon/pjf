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

type ResolvedCommand struct {
	Name      string
	Command   string
	IsProject bool
}

func (s *ProjectService) ResolveCommands(p project.Project) []ResolvedCommand {
	var projectCmds []config.CommandDef
	for _, pc := range s.Cfg.ProjectCommands {
		expanded := pc.Path
		if home, err := os.UserHomeDir(); err == nil && len(expanded) > 0 && expanded[0] == '~' {
			expanded = home + expanded[1:]
		}
		if expanded == p.Path {
			projectCmds = pc.Commands
			break
		}
	}

	projectNames := make(map[string]bool)
	var result []ResolvedCommand
	for _, cmd := range projectCmds {
		result = append(result, ResolvedCommand{Name: cmd.Name, Command: cmd.Command, IsProject: true})
		projectNames[cmd.Name] = true
	}
	for _, cmd := range s.Cfg.GlobalCommands {
		if !projectNames[cmd.Name] {
			result = append(result, ResolvedCommand{Name: cmd.Name, Command: cmd.Command, IsProject: false})
		}
	}
	return result
}

func (s *ProjectService) RunCommand(p project.Project, command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = p.Path
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (s *ProjectService) RemoveProjectIDE(p project.Project) {
	delete(s.Cfg.ProjectIDEs, p.Path)
}

func (s *ProjectService) RemoveSavedCommand(p project.Project, index int) {
	for i, pc := range s.Cfg.ProjectCommands {
		expanded := pc.Path
		if home, err := os.UserHomeDir(); err == nil && len(expanded) > 0 && expanded[0] == '~' {
			expanded = home + expanded[1:]
		}
		if expanded == p.Path {
			if index >= 0 && index < len(pc.Commands) {
				s.Cfg.ProjectCommands[i].Commands = append(pc.Commands[:index], pc.Commands[index+1:]...)
			}
			return
		}
	}
}

func (s *ProjectService) AddGlobalCommand(name, command string) {
	s.Cfg.GlobalCommands = append(s.Cfg.GlobalCommands, config.CommandDef{Name: name, Command: command})
}

func (s *ProjectService) RemoveGlobalCommand(index int) {
	if index >= 0 && index < len(s.Cfg.GlobalCommands) {
		s.Cfg.GlobalCommands = append(s.Cfg.GlobalCommands[:index], s.Cfg.GlobalCommands[index+1:]...)
	}
}

func (s *ProjectService) AddScanDir(dir string) {
	for _, d := range s.Cfg.ScanDirs {
		if d == dir {
			return
		}
	}
	s.Cfg.ScanDirs = append(s.Cfg.ScanDirs, dir)
}

func (s *ProjectService) RemoveScanDir(index int) {
	if index >= 0 && index < len(s.Cfg.ScanDirs) && len(s.Cfg.ScanDirs) > 1 {
		s.Cfg.ScanDirs = append(s.Cfg.ScanDirs[:index], s.Cfg.ScanDirs[index+1:]...)
	}
}

func (s *ProjectService) AddExclude(dir string) {
	s.Cfg.ExtraExcludes = append(s.Cfg.ExtraExcludes, dir)
}

func (s *ProjectService) RemoveExclude(index int) {
	if index >= 0 && index < len(s.Cfg.ExtraExcludes) {
		s.Cfg.ExtraExcludes = append(s.Cfg.ExtraExcludes[:index], s.Cfg.ExtraExcludes[index+1:]...)
	}
}

func (s *ProjectService) SetMaxDepth(depth int) {
	s.Cfg.MaxDepth = depth
}

func (s *ProjectService) SetCacheTTL(hours int) {
	s.Cfg.CacheTTL = hours
}

func (s *ProjectService) SetDefaultIDE(projectType, ideSlug string) {
	if s.Cfg.DefaultIDEs == nil {
		s.Cfg.DefaultIDEs = make(map[string]string)
	}
	s.Cfg.DefaultIDEs[projectType] = ideSlug
}

func (s *ProjectService) RemoveDefaultIDE(projectType string) {
	delete(s.Cfg.DefaultIDEs, projectType)
}

type GroupedSection struct {
	Name      string
	Collapsed bool
	IsOther   bool
	Projects  []project.Project
}

func (s *ProjectService) ProjectGroup(p project.Project) string {
	for _, g := range s.Cfg.Groups {
		for _, path := range g.Projects {
			if path == p.Path {
				return g.Name
			}
		}
	}
	return ""
}

func (s *ProjectService) SetProjectGroup(p project.Project, groupName string) {
	s.RemoveProjectGroup(p)
	for i, g := range s.Cfg.Groups {
		if g.Name == groupName {
			s.Cfg.Groups[i].Projects = append(s.Cfg.Groups[i].Projects, p.Path)
			return
		}
	}
}

func (s *ProjectService) RemoveProjectGroup(p project.Project) {
	for i, g := range s.Cfg.Groups {
		for j, path := range g.Projects {
			if path == p.Path {
				s.Cfg.Groups[i].Projects = append(g.Projects[:j], g.Projects[j+1:]...)
				return
			}
		}
	}
}

func (s *ProjectService) AddGroup(name string) {
	for _, g := range s.Cfg.Groups {
		if g.Name == name {
			return
		}
	}
	s.Cfg.Groups = append(s.Cfg.Groups, config.GroupDef{Name: name})
}

func (s *ProjectService) RemoveGroup(name string) {
	for i, g := range s.Cfg.Groups {
		if g.Name == name {
			s.Cfg.Groups = append(s.Cfg.Groups[:i], s.Cfg.Groups[i+1:]...)
			return
		}
	}
}

func (s *ProjectService) RenameGroup(oldName, newName string) {
	for i, g := range s.Cfg.Groups {
		if g.Name == oldName {
			s.Cfg.Groups[i].Name = newName
			return
		}
	}
}

func (s *ProjectService) SetGroupCollapsed(name string, collapsed bool) {
	for i, g := range s.Cfg.Groups {
		if g.Name == name {
			s.Cfg.Groups[i].Collapsed = collapsed
			return
		}
	}
}

func (s *ProjectService) GroupedProjects(projects []project.Project) []GroupedSection {
	assigned := make(map[string]bool)
	for _, g := range s.Cfg.Groups {
		for _, path := range g.Projects {
			assigned[path] = true
		}
	}

	grouped := make(map[string][]project.Project)
	for _, p := range projects {
		groupName := s.ProjectGroup(p)
		if groupName != "" {
			grouped[groupName] = append(grouped[groupName], p)
		}
	}

	var sections []GroupedSection
	for _, g := range s.Cfg.Groups {
		sections = append(sections, GroupedSection{
			Name:      g.Name,
			Collapsed: g.Collapsed,
			Projects:  grouped[g.Name],
		})
	}

	var other []project.Project
	for _, p := range projects {
		if !assigned[p.Path] {
			other = append(other, p)
		}
	}
	if len(other) > 0 || len(s.Cfg.Groups) > 0 {
		sections = append(sections, GroupedSection{
			Name:     "Other",
			IsOther:  true,
			Projects: other,
		})
	}

	return sections
}

func (s *ProjectService) SaveCommand(p project.Project, command string) {
	for i, pc := range s.Cfg.ProjectCommands {
		expanded := pc.Path
		if home, err := os.UserHomeDir(); err == nil && len(expanded) > 0 && expanded[0] == '~' {
			expanded = home + expanded[1:]
		}
		if expanded == p.Path {
			s.Cfg.ProjectCommands[i].Commands = append(s.Cfg.ProjectCommands[i].Commands, config.CommandDef{
				Name:    command,
				Command: command,
			})
			return
		}
	}
	s.Cfg.ProjectCommands = append(s.Cfg.ProjectCommands, config.ProjectCommandSet{
		Path: p.Path,
		Commands: []config.CommandDef{
			{Name: command, Command: command},
		},
	})
}

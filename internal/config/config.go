package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type CommandDef struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

type ProjectCommandSet struct {
	Path     string       `yaml:"path"`
	Commands []CommandDef `yaml:"commands"`
}

type Config struct {
	ScanDirs        []string            `yaml:"scan_dirs"`
	ExtraExcludes   []string            `yaml:"extra_excludes"`
	ManualProjects  []string            `yaml:"manual_projects"`
	HiddenProjects  []string            `yaml:"hidden_projects"`
	DefaultIDEs     map[string]string   `yaml:"default_ides"`
	ProjectIDEs     map[string]string   `yaml:"project_ides"`
	GlobalCommands  []CommandDef        `yaml:"global_commands"`
	ProjectCommands []ProjectCommandSet `yaml:"project_commands"`
	MaxDepth        int                 `yaml:"max_depth"`
	CacheTTL        int                 `yaml:"cache_ttl"`
}

func Load(path string) (*Config, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return defaultConfig(), true, nil
		}
		return nil, false, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, false, err
	}
	applyDefaults(&cfg)
	return &cfg, false, nil
}

func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pjf", "config.yaml")
}

func DefaultCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "pjf", "projects.json")
}

func defaultConfig() *Config {
	return &Config{
		MaxDepth: 5,
		CacheTTL: 24,
	}
}

func applyDefaults(cfg *Config) {
	if cfg.MaxDepth == 0 {
		cfg.MaxDepth = 5
	}
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 24
	}
}

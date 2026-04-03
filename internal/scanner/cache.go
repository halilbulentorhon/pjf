package scanner

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/halilbulentorhon/pjf/internal/project"
)

type Cache struct {
	Projects []project.Project `json:"projects"`
	LastScan time.Time         `json:"last_scan"`
}

func (c *Cache) IsStale(ttlHours int) bool {
	return time.Since(c.LastScan) > time.Duration(ttlHours)*time.Hour
}

type JSONCacheStore struct {
	Path string
}

func (s *JSONCacheStore) Load() (*Cache, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var c Cache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *JSONCacheStore) Save(projects []project.Project) error {
	c := Cache{
		Projects: projects,
		LastScan: time.Now(),
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0755); err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0644)
}

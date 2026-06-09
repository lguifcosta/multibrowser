package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	BrowserPath       string `json:"browser_path"`
	BrowserName       string `json:"browser_name"`
	ShowUnassignedTab bool   `json:"show_unassigned_tab"`
	DefaultTabID      string `json:"default_tab_id"`
}

type Manager struct {
	path   string
	config Config
	mu     sync.Mutex
}

func NewManager(baseDir string) (*Manager, error) {
	path := filepath.Join(baseDir, "config.json")

	m := &Manager{path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &m.config); err != nil {
		return m, nil
	}

	return m, nil
}

func (m *Manager) Get() Config {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.config
}

func (m *Manager) SetBrowser(name, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.BrowserPath = path
	m.config.BrowserName = name
	return m.save()
}

func (m *Manager) UpdateConfig(cfg Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = cfg
	return m.save()
}

func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, data, 0644)
}

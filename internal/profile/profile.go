package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
)

type ProfileFlags struct {
	RestoreLastSession               bool   `json:"restore_last_session"`
	UserAgent                        string `json:"user_agent"`
	Lang                             string `json:"lang"`
	WindowSize                       string `json:"window_size"`
	ProxyServer                      string `json:"proxy_server"`
	ProxyBypassList                  string `json:"proxy_bypass_list"`
	DisableBackgroundNetworking      bool   `json:"disable_background_networking"`
	DisableBackgroundTimerThrottling bool   `json:"disable_background_timer_throttling"`
	DisableRendererBackgrounding     bool   `json:"disable_renderer_backgrounding"`
	DisableFeaturesTranslateUI       bool   `json:"disable_features_translate_ui"`
	DisableExtensions                bool   `json:"disable_extensions"`
	DisableSync                      bool   `json:"disable_sync"`
}

type Profile struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	CreatedAt time.Time    `json:"created_at"`
	LastUsed  time.Time    `json:"last_used"`
	Status    Status       `json:"status"`
	Flags     ProfileFlags `json:"flags"`
}

type Metadata struct {
	Profiles []Profile `json:"profiles"`
}

type Manager struct {
	baseDir      string
	metadataPath string
	mu           sync.Mutex
}

func NewManager(baseDir string) (*Manager, error) {
	profilesDir := filepath.Join(baseDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create profiles directory: %w", err)
	}

	backupsDir := filepath.Join(baseDir, "backups")
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backups directory: %w", err)
	}

	m := &Manager{
		baseDir:      baseDir,
		metadataPath: filepath.Join(baseDir, "metadata.json"),
	}

	if _, err := os.Stat(m.metadataPath); os.IsNotExist(err) {
		if err := m.saveMetadata(&Metadata{Profiles: []Profile{}}); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *Manager) BaseDir() string {
	return m.baseDir
}

func (m *Manager) ProfileDir(id string) string {
	return filepath.Join(m.baseDir, "profiles", id)
}

func (m *Manager) BackupsDir() string {
	return filepath.Join(m.baseDir, "backups")
}

func (m *Manager) loadMetadata() (*Metadata, error) {
	data, err := os.ReadFile(m.metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &meta, nil
}

func (m *Manager) saveMetadata(meta *Metadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(m.metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func (m *Manager) Create(name string) (*Profile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta, err := m.loadMetadata()
	if err != nil {
		return nil, err
	}

	profile := Profile{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		Status:    StatusStopped,
	}

	profileDir := m.ProfileDir(profile.ID)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create profile directory: %w", err)
	}

	meta.Profiles = append(meta.Profiles, profile)
	if err := m.saveMetadata(meta); err != nil {
		os.RemoveAll(profileDir)
		return nil, err
	}

	return &profile, nil
}

func (m *Manager) List() ([]Profile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta, err := m.loadMetadata()
	if err != nil {
		return nil, err
	}

	return meta.Profiles, nil
}

func (m *Manager) Get(id string) (*Profile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta, err := m.loadMetadata()
	if err != nil {
		return nil, err
	}

	for _, p := range meta.Profiles {
		if p.ID == id {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("profile not found: %s", id)
}

func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta, err := m.loadMetadata()
	if err != nil {
		return err
	}

	idx := -1
	for i, p := range meta.Profiles {
		if p.ID == id {
			if p.Status == StatusRunning {
				return fmt.Errorf("cannot delete running profile: %s", id)
			}
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("profile not found: %s", id)
	}

	profileDir := m.ProfileDir(id)
	if err := os.RemoveAll(profileDir); err != nil {
		return fmt.Errorf("failed to remove profile directory: %w", err)
	}

	meta.Profiles = append(meta.Profiles[:idx], meta.Profiles[idx+1:]...)
	return m.saveMetadata(meta)
}

func (m *Manager) Rename(id, newName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta, err := m.loadMetadata()
	if err != nil {
		return err
	}

	for i, p := range meta.Profiles {
		if p.ID == id {
			meta.Profiles[i].Name = newName
			return m.saveMetadata(meta)
		}
	}

	return fmt.Errorf("profile not found: %s", id)
}

func (m *Manager) Clone(id, newName string) (*Profile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta, err := m.loadMetadata()
	if err != nil {
		return nil, err
	}

	var source *Profile
	for _, p := range meta.Profiles {
		if p.ID == id {
			if p.Status == StatusRunning {
				return nil, fmt.Errorf("cannot clone running profile: %s", id)
			}
			source = &p
			break
		}
	}

	if source == nil {
		return nil, fmt.Errorf("profile not found: %s", id)
	}

	newProfile := Profile{
		ID:        uuid.New().String(),
		Name:      newName,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		Status:    StatusStopped,
	}

	srcDir := m.ProfileDir(source.ID)
	dstDir := m.ProfileDir(newProfile.ID)

	if err := copyDir(srcDir, dstDir); err != nil {
		return nil, fmt.Errorf("failed to clone profile directory: %w", err)
	}

	meta.Profiles = append(meta.Profiles, newProfile)
	if err := m.saveMetadata(meta); err != nil {
		os.RemoveAll(dstDir)
		return nil, err
	}

	return &newProfile, nil
}

func (m *Manager) UpdateFlags(id string, flags ProfileFlags) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta, err := m.loadMetadata()
	if err != nil {
		return err
	}

	for i, p := range meta.Profiles {
		if p.ID == id {
			meta.Profiles[i].Flags = flags
			return m.saveMetadata(meta)
		}
	}

	return fmt.Errorf("profile not found: %s", id)
}

func (m *Manager) UpdateStatus(id string, status Status) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta, err := m.loadMetadata()
	if err != nil {
		return err
	}

	for i, p := range meta.Profiles {
		if p.ID == id {
			meta.Profiles[i].Status = status
			if status == StatusRunning {
				meta.Profiles[i].LastUsed = time.Now()
			}
			return m.saveMetadata(meta)
		}
	}

	return fmt.Errorf("profile not found: %s", id)
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.Name() == ".lock" {
			continue
		}

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, info.Mode()); err != nil {
				return err
			}
		}
	}

	return nil
}

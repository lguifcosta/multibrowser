package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func lockPath(profileDir string) string {
	return filepath.Join(profileDir, ".lock")
}

func (m *Manager) Acquire(profileDir string, pid int) error {
	lp := lockPath(profileDir)

	if m.IsLocked(profileDir) {
		return fmt.Errorf("profile is already locked: %s", profileDir)
	}

	return os.WriteFile(lp, []byte(strconv.Itoa(pid)), 0644)
}

func (m *Manager) Release(profileDir string) error {
	lp := lockPath(profileDir)

	if _, err := os.Stat(lp); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(lp)
}

func (m *Manager) IsLocked(profileDir string) bool {
	lp := lockPath(profileDir)

	data, err := os.ReadFile(lp)
	if err != nil {
		return false
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		// Lock file corrupted, clean it up
		os.Remove(lp)
		return false
	}

	if !processExists(pid) {
		// Orphaned lock, clean it up
		os.Remove(lp)
		return false
	}

	return true
}

func (m *Manager) GetPID(profileDir string) (int, error) {
	lp := lockPath(profileDir)

	data, err := os.ReadFile(lp)
	if err != nil {
		return 0, fmt.Errorf("no lock file: %w", err)
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("invalid PID in lock file: %w", err)
	}

	return pid, nil
}

func processExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

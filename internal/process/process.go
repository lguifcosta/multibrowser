package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"multibrowser/internal/lock"
	"multibrowser/internal/logger"
	"multibrowser/internal/profile"
)

type Manager struct {
	profileManager *profile.Manager
	lockManager    *lock.Manager
	chromiumPath   string
}

func NewManager(pm *profile.Manager, lm *lock.Manager) (*Manager, error) {
	chromiumPath, err := detectChromium()
	if err != nil {
		return nil, err
	}

	return &Manager{
		profileManager: pm,
		lockManager:    lm,
		chromiumPath:   chromiumPath,
	}, nil
}

func (m *Manager) SetChromiumPath(path string) {
	m.chromiumPath = path
}

func (m *Manager) GetChromiumPath() string {
	return m.chromiumPath
}

func (m *Manager) Launch(profileID string) (int, error) {
	p, err := m.profileManager.Get(profileID)
	if err != nil {
		return 0, err
	}

	profileDir := m.profileManager.ProfileDir(p.ID)

	if m.lockManager.IsLocked(profileDir) {
		return 0, fmt.Errorf("profile is already running: %s", p.Name)
	}

	cmd := exec.Command(m.chromiumPath,
		"--user-data-dir="+profileDir,
		"--no-first-run",
		"--no-default-browser-check",
	)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start chromium: %w", err)
	}

	pid := cmd.Process.Pid

	if err := m.lockManager.Acquire(profileDir, pid); err != nil {
		cmd.Process.Kill()
		return 0, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if err := m.profileManager.UpdateStatus(profileID, profile.StatusRunning); err != nil {
		if logger.Error != nil {
			logger.Error.Printf("failed to update profile status: %v", err)
		}
	}

	if logger.Info != nil {
		logger.Info.Printf("launched chromium for profile %s (PID: %d)", p.Name, pid)
	}

	go m.waitForExit(cmd, profileID, profileDir)

	return pid, nil
}

func (m *Manager) Stop(profileID string) error {
	profileDir := m.profileManager.ProfileDir(profileID)

	pid, err := m.lockManager.GetPID(profileDir)
	if err != nil {
		return fmt.Errorf("profile is not running: %s", profileID)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		m.lockManager.Release(profileDir)
		m.profileManager.UpdateStatus(profileID, profile.StatusStopped)
		return nil
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		process.Kill()
	}

	m.lockManager.Release(profileDir)
	m.profileManager.UpdateStatus(profileID, profile.StatusStopped)

	if logger.Info != nil {
		logger.Info.Printf("stopped chromium for profile %s (PID: %d)", profileID, pid)
	}

	return nil
}

func (m *Manager) IsRunning(profileID string) bool {
	profileDir := m.profileManager.ProfileDir(profileID)
	return m.lockManager.IsLocked(profileDir)
}

func (m *Manager) waitForExit(cmd *exec.Cmd, profileID, profileDir string) {
	cmd.Wait()
	m.lockManager.Release(profileDir)
	m.profileManager.UpdateStatus(profileID, profile.StatusStopped)

	if logger.Info != nil {
		logger.Info.Printf("chromium exited for profile %s", profileID)
	}
}

func detectChromium() (string, error) {
	var candidates []string

	if runtime.GOOS == "windows" {
		programFiles := os.Getenv("PROGRAMFILES")
		programFilesX86 := os.Getenv("PROGRAMFILES(X86)")
		localAppData := os.Getenv("LOCALAPPDATA")

		candidates = []string{
			filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(programFilesX86, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(localAppData, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(programFiles, "Chromium", "Application", "chrome.exe"),
		}
	} else {
		candidates = []string{
			"chromium",
			"chromium-browser",
			"google-chrome",
			"google-chrome-stable",
		}
	}

	for _, candidate := range candidates {
		if runtime.GOOS == "windows" {
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		} else {
			if path, err := exec.LookPath(candidate); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("chromium not found. Searched: %s", strings.Join(candidates, ", "))
}

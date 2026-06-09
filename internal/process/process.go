package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"multibrowser/internal/config"
	"multibrowser/internal/lock"
	"multibrowser/internal/logger"
	"multibrowser/internal/profile"
)

type Browser struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type ProcessTelemetry struct {
	PID         int     `json:"pid"`
	CPUPercent  float64 `json:"cpu_percent"`
	RAMMB       float64 `json:"ram_mb"`
	LifetimeSec int64   `json:"lifetime_sec"`
	IsRunning   bool    `json:"is_running"`
}

type runningProcess struct {
	pid           int
	startedAt     time.Time
	cmd           *exec.Cmd
	prevProcJiff  uint64
	prevTotalJiff uint64
	cpuPercent    float64
	ramMB         float64
}

type Manager struct {
	profileManager *profile.Manager
	lockManager    *lock.Manager
	configManager  *config.Manager
	browserPath    string
	browserName    string
	running        map[string]*runningProcess
	runMu          sync.RWMutex
}

func NewManager(pm *profile.Manager, lm *lock.Manager, cm *config.Manager) *Manager {
	m := &Manager{
		profileManager: pm,
		lockManager:    lm,
		configManager:  cm,
		running:        make(map[string]*runningProcess),
	}

	cfg := cm.Get()
	if cfg.BrowserPath != "" {
		if _, err := os.Stat(cfg.BrowserPath); err == nil {
			m.browserPath = cfg.BrowserPath
			m.browserName = cfg.BrowserName
		} else if path, err := exec.LookPath(cfg.BrowserPath); err == nil {
			m.browserPath = path
			m.browserName = cfg.BrowserName
		}
	}

	if m.browserPath == "" {
		browsers := DetectBrowsers()
		if len(browsers) > 0 {
			m.browserPath = browsers[0].Path
			m.browserName = browsers[0].Name
		}
	}

	go m.monitorTelemetry()

	return m
}

func (m *Manager) monitorTelemetry() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		m.runMu.Lock()
		for _, rp := range m.running {
			cpu, ram, curProc, curTotal := sampleTelemetry(rp.pid, rp.prevProcJiff, rp.prevTotalJiff)
			rp.cpuPercent = cpu
			rp.ramMB = ram
			rp.prevProcJiff = curProc
			rp.prevTotalJiff = curTotal
		}
		m.runMu.Unlock()
	}
}

func (m *Manager) GetTelemetry(profileID string) (ProcessTelemetry, bool) {
	m.runMu.RLock()
	defer m.runMu.RUnlock()

	rp, ok := m.running[profileID]
	if !ok {
		return ProcessTelemetry{IsRunning: false}, false
	}

	return ProcessTelemetry{
		PID:         rp.pid,
		CPUPercent:  rp.cpuPercent,
		RAMMB:       rp.ramMB,
		LifetimeSec: int64(time.Since(rp.startedAt).Seconds()),
		IsRunning:   true,
	}, true
}

func (m *Manager) SetBrowser(name, path string) error {
	m.browserPath = path
	m.browserName = name
	return m.configManager.SetBrowser(name, path)
}

func (m *Manager) GetBrowserPath() string {
	return m.browserPath
}

func (m *Manager) GetBrowserName() string {
	return m.browserName
}

func (m *Manager) HasBrowser() bool {
	return m.browserPath != ""
}

func (m *Manager) Launch(profileID string) (int, error) {
	if m.browserPath == "" {
		return 0, fmt.Errorf("no browser configured. Please select a browser in settings")
	}

	p, err := m.profileManager.Get(profileID)
	if err != nil {
		return 0, err
	}

	profileDir := m.profileManager.ProfileDir(p.ID)

	if m.lockManager.IsLocked(profileDir) {
		return 0, fmt.Errorf("profile is already running: %s", p.Name)
	}

	globalFlags := m.configManager.Get().GlobalFlags
	args := buildArgs(profileDir, p.Flags, globalFlags)
	cmd := exec.Command(m.browserPath, args...)
	setSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start browser: %w", err)
	}

	pid := cmd.Process.Pid

	if err := m.lockManager.Acquire(profileDir, pid); err != nil {
		cmd.Process.Kill()
		return 0, fmt.Errorf("failed to acquire lock: %w", err)
	}

	m.runMu.Lock()
	m.running[profileID] = &runningProcess{
		pid:       pid,
		startedAt: time.Now(),
		cmd:       cmd,
	}
	m.runMu.Unlock()

	if err := m.profileManager.UpdateStatus(profileID, profile.StatusRunning); err != nil {
		if logger.Error != nil {
			logger.Error.Printf("failed to update profile status: %v", err)
		}
	}

	if logger.Info != nil {
		logger.Info.Printf("launched %s for profile %s (PID: %d)", m.browserName, p.Name, pid)
	}

	go m.waitForExit(cmd, profileID, profileDir)

	return pid, nil
}

func buildArgs(profileDir string, flags profile.ProfileFlags, global config.GlobalFlags) []string {
	args := []string{
		"--user-data-dir=" + profileDir,
		"--no-first-run",
		"--no-default-browser-check",
	}

	resolveBool := func(pFlag *bool, gFlag *bool) bool {
		if pFlag != nil {
			return *pFlag
		}
		if gFlag != nil {
			return *gFlag
		}
		return false
	}

	if resolveBool(flags.RestoreLastSession, global.RestoreLastSession) {
		args = append(args, "--restore-last-session")
	}
	if flags.UserAgent != "" {
		args = append(args, "--user-agent="+flags.UserAgent)
	}
	if flags.Lang != "" {
		args = append(args, "--lang="+flags.Lang)
	}
	if flags.WindowSize != "" {
		args = append(args, "--window-size="+flags.WindowSize)
	}
	if flags.ProxyServer != "" {
		args = append(args, "--proxy-server="+flags.ProxyServer)
		if flags.ProxyBypassList != "" {
			args = append(args, "--proxy-bypass-list="+flags.ProxyBypassList)
		}
	}
	if resolveBool(flags.DisableBackgroundNetworking, global.DisableBackgroundNetworking) {
		args = append(args, "--disable-background-networking")
	}
	if resolveBool(flags.DisableBackgroundTimerThrottling, global.DisableBackgroundTimerThrottling) {
		args = append(args, "--disable-background-timer-throttling")
	}
	if resolveBool(flags.DisableRendererBackgrounding, global.DisableRendererBackgrounding) {
		args = append(args, "--disable-renderer-backgrounding")
	}
	if resolveBool(flags.DisableFeaturesTranslateUI, global.DisableFeaturesTranslateUI) {
		args = append(args, "--disable-features=TranslateUI")
	}
	if resolveBool(flags.DisableExtensions, global.DisableExtensions) {
		args = append(args, "--disable-extensions")
	}
	if resolveBool(flags.DisableSync, global.DisableSync) {
		args = append(args, "--disable-sync")
	}
	return args
}

func (m *Manager) Stop(profileID string) error {
	profileDir := m.profileManager.ProfileDir(profileID)

	pid, err := m.lockManager.GetPID(profileDir)
	if err != nil {
		return fmt.Errorf("profile is not running: %s", profileID)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		m.lockManager.Release(profileDir)
		m.profileManager.UpdateStatus(profileID, profile.StatusStopped)
		m.runMu.Lock()
		delete(m.running, profileID)
		m.runMu.Unlock()
		return nil
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		proc.Kill()
	}

	m.runMu.Lock()
	delete(m.running, profileID)
	m.runMu.Unlock()

	m.lockManager.Release(profileDir)
	m.profileManager.UpdateStatus(profileID, profile.StatusStopped)

	if logger.Info != nil {
		logger.Info.Printf("stopped browser for profile %s (PID: %d)", profileID, pid)
	}

	return nil
}

func (m *Manager) IsRunning(profileID string) bool {
	profileDir := m.profileManager.ProfileDir(profileID)
	return m.lockManager.IsLocked(profileDir)
}

func (m *Manager) waitForExit(cmd *exec.Cmd, profileID, profileDir string) {
	cmd.Wait()

	m.runMu.Lock()
	delete(m.running, profileID)
	m.runMu.Unlock()

	m.lockManager.Release(profileDir)
	m.profileManager.UpdateStatus(profileID, profile.StatusStopped)

	if logger.Info != nil {
		logger.Info.Printf("browser exited for profile %s", profileID)
	}
}

func DetectBrowsers() []Browser {
	type candidate struct {
		Name string
		Bins []string
	}

	var found []Browser

	if runtime.GOOS == "windows" {
		programFiles := os.Getenv("PROGRAMFILES")
		programFilesX86 := os.Getenv("PROGRAMFILES(X86)")
		localAppData := os.Getenv("LOCALAPPDATA")

		candidates := []candidate{
			{"Google Chrome", []string{
				filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe"),
				filepath.Join(programFilesX86, "Google", "Chrome", "Application", "chrome.exe"),
				filepath.Join(localAppData, "Google", "Chrome", "Application", "chrome.exe"),
			}},
			{"Chromium", []string{
				filepath.Join(programFiles, "Chromium", "Application", "chrome.exe"),
				filepath.Join(localAppData, "Chromium", "Application", "chrome.exe"),
			}},
			{"Brave", []string{
				filepath.Join(programFiles, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
				filepath.Join(programFilesX86, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
				filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
			}},
			{"Microsoft Edge", []string{
				filepath.Join(programFiles, "Microsoft", "Edge", "Application", "msedge.exe"),
				filepath.Join(programFilesX86, "Microsoft", "Edge", "Application", "msedge.exe"),
			}},
			{"Vivaldi", []string{
				filepath.Join(localAppData, "Vivaldi", "Application", "vivaldi.exe"),
				filepath.Join(programFiles, "Vivaldi", "Application", "vivaldi.exe"),
			}},
			{"Opera", []string{
				filepath.Join(localAppData, "Programs", "Opera", "opera.exe"),
				filepath.Join(programFiles, "Opera", "opera.exe"),
			}},
			{"Opera GX", []string{
				filepath.Join(localAppData, "Programs", "Opera GX", "opera.exe"),
			}},
		}

		for _, c := range candidates {
			for _, bin := range c.Bins {
				if _, err := os.Stat(bin); err == nil {
					found = append(found, Browser{Name: c.Name, Path: bin})
					break
				}
			}
		}
	} else {
		candidates := []candidate{
			{"Google Chrome", []string{"google-chrome", "google-chrome-stable"}},
			{"Chromium", []string{"chromium", "chromium-browser"}},
			{"Brave", []string{"brave-browser", "brave-browser-stable", "brave"}},
			{"Microsoft Edge", []string{"microsoft-edge", "microsoft-edge-stable"}},
			{"Vivaldi", []string{"vivaldi", "vivaldi-stable"}},
			{"Opera", []string{"opera"}},
		}

		for _, c := range candidates {
			for _, bin := range c.Bins {
				if path, err := exec.LookPath(bin); err == nil {
					found = append(found, Browser{Name: c.Name, Path: path})
					break
				}
			}
		}
	}

	return found
}

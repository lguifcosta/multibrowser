package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"multibrowser/internal/backup"
	"multibrowser/internal/cache"
	"multibrowser/internal/config"
	"multibrowser/internal/lock"
	"multibrowser/internal/logger"
	"multibrowser/internal/process"
	"multibrowser/internal/profile"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx            context.Context
	profileManager *profile.Manager
	lockManager    *lock.Manager
	processManager *process.Manager
	backupService  *backup.Service
	cacheCleaner   *cache.Cleaner
	configManager  *config.Manager
}

type ProfileInfo struct {
	ProfileDir  string  `json:"profile_dir"`
	DiskUsageMB float64 `json:"disk_usage_mb"`
	CrashCount  int     `json:"crash_count"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	homeDir, err := os.UserHomeDir()
	if err != nil {
		a.criticalError("Failed to get home directory", err)
		return
	}

	baseDir := filepath.Join(homeDir, ".multibrowser")

	if err := logger.Init(baseDir); err != nil {
		a.criticalError("Failed to initialize logger", err)
		return
	}

	a.configManager, err = config.NewManager(baseDir)
	if err != nil {
		a.criticalError("Failed to initialize config manager", err)
		return
	}

	a.lockManager = lock.NewManager()

	a.profileManager, err = profile.NewManager(baseDir)
	if err != nil {
		a.criticalError("Failed to initialize profile manager", err)
		return
	}

	a.processManager = process.NewManager(a.profileManager, a.lockManager, a.configManager)

	a.backupService = backup.NewService(a.profileManager, a.lockManager)
	a.cacheCleaner = cache.NewCleaner(a.profileManager, a.lockManager)

	logger.Info.Println("application started")
}

func (a *App) criticalError(title string, err error) {
	msg := fmt.Sprintf("%s: %v", title, err)
	if logger.Error != nil {
		logger.Error.Println(msg)
	} else {
		fmt.Printf("CRITICAL ERROR: %s\n", msg)
	}

	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.ErrorDialog,
		Title:   "Critical Error",
		Message: msg,
	})
	runtime.Quit(a.ctx)
}

// Profile methods

func (a *App) CreateProfile(name string) (*profile.Profile, error) {
	return a.profileManager.Create(name)
}

func (a *App) ListProfiles() ([]profile.Profile, error) {
	return a.profileManager.List()
}

func (a *App) CreateGroup(name string) (*profile.Group, error) {
	return a.profileManager.CreateGroup(name)
}

func (a *App) ListGroups() ([]profile.Group, error) {
	return a.profileManager.ListGroups()
}

func (a *App) RenameGroup(id, newName string) error {
	return a.profileManager.RenameGroup(id, newName)
}

func (a *App) DeleteGroup(id string) error {
	return a.profileManager.DeleteGroup(id)
}

func (a *App) ReorderGroups(ids []string) error {
	return a.profileManager.ReorderGroups(ids)
}

func (a *App) AssignProfileToGroup(profileID, groupID string) error {
	return a.profileManager.AssignToGroup(profileID, groupID)
}

func (a *App) DeleteProfile(id string) error {
	return a.profileManager.Delete(id)
}

func (a *App) RenameProfile(id, newName string) error {
	return a.profileManager.Rename(id, newName)
}

func (a *App) CloneProfile(id, newName string) (*profile.Profile, error) {
	return a.profileManager.Clone(id, newName)
}

func (a *App) GetProfileFlags(id string) (*profile.ProfileFlags, error) {
	p, err := a.profileManager.Get(id)
	if err != nil {
		return nil, err
	}
	return &p.Flags, nil
}

func (a *App) UpdateProfileFlags(id string, flags profile.ProfileFlags) error {
	return a.profileManager.UpdateFlags(id, flags)
}

func (a *App) GetProfileTelemetry(id string) (*process.ProcessTelemetry, error) {
	telem, ok := a.processManager.GetTelemetry(id)
	if !ok {
		return &process.ProcessTelemetry{IsRunning: false}, nil
	}
	return &telem, nil
}

func (a *App) GetProfileInfo(id string) (*ProfileInfo, error) {
	profileDir := a.profileManager.ProfileDir(id)
	return &ProfileInfo{
		ProfileDir:  profileDir,
		DiskUsageMB: dirSizeMB(profileDir),
		CrashCount:  countCrashReports(profileDir),
	}, nil
}

// Process methods

func (a *App) LaunchProfile(id string) (int, error) {
	return a.processManager.Launch(id)
}

func (a *App) StopProfile(id string) error {
	return a.processManager.Stop(id)
}

func (a *App) IsProfileRunning(id string) bool {
	return a.processManager.IsRunning(id)
}

// Browser methods

func (a *App) GetBrowserPath() string {
	return a.processManager.GetBrowserPath()
}

func (a *App) GetBrowserName() string {
	return a.processManager.GetBrowserName()
}

func (a *App) HasBrowser() bool {
	return a.processManager.HasBrowser()
}

func (a *App) DetectBrowsers() []process.Browser {
	return process.DetectBrowsers()
}

func (a *App) SetBrowser(name, path string) error {
	return a.processManager.SetBrowser(name, path)
}

func (a *App) SetCustomBrowserPath(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("executable not found: %s", path)
	}
	return a.processManager.SetBrowser("Custom", path)
}

// Backup methods

func (a *App) ExportBackup(profileID string, excludeCache bool, password string) (string, error) {
	return a.backupService.Export(profileID, excludeCache, password)
}

func (a *App) ImportBackup(backupPath, newName, password string) (*profile.Profile, error) {
	return a.backupService.Import(backupPath, newName, password)
}

// Cache methods

func (a *App) CleanCache(profileID string) (int64, error) {
	return a.cacheCleaner.Clean(profileID)
}

func (a *App) GetConfig() config.Config {
	return a.configManager.Get()
}

func (a *App) UpdateConfig(cfg config.Config) error {
	return a.configManager.UpdateConfig(cfg)
}

// Helpers

func dirSizeMB(path string) float64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		size += info.Size()
		return nil
	})
	return float64(size) / 1024.0 / 1024.0
}

func countCrashReports(profileDir string) int {
	reportDir := filepath.Join(profileDir, "Crashpad", "reports")
	entries, err := os.ReadDir(reportDir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			count++
		}
	}
	return count
}

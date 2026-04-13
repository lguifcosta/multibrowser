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

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("failed to get home directory: " + err.Error())
	}

	baseDir := filepath.Join(homeDir, ".multibrowser")

	if err := logger.Init(baseDir); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	a.configManager, err = config.NewManager(baseDir)
	if err != nil {
		panic("failed to initialize config manager: " + err.Error())
	}

	a.lockManager = lock.NewManager()

	a.profileManager, err = profile.NewManager(baseDir)
	if err != nil {
		panic("failed to initialize profile manager: " + err.Error())
	}

	a.processManager = process.NewManager(a.profileManager, a.lockManager, a.configManager)

	a.backupService = backup.NewService(a.profileManager, a.lockManager)
	a.cacheCleaner = cache.NewCleaner(a.profileManager, a.lockManager)

	logger.Info.Println("application started")
}

// Profile methods

func (a *App) CreateProfile(name string) (*profile.Profile, error) {
	return a.profileManager.Create(name)
}

func (a *App) ListProfiles() ([]profile.Profile, error) {
	return a.profileManager.List()
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

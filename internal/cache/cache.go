package cache

import (
	"fmt"
	"os"
	"path/filepath"

	"multibrowser/internal/lock"
	"multibrowser/internal/logger"
	"multibrowser/internal/profile"
)

var targetDirs = []string{
	"Cache",
	"Code Cache",
	"GPUCache",
	"Crashpad",
}

type Cleaner struct {
	profileManager *profile.Manager
	lockManager    *lock.Manager
}

func NewCleaner(pm *profile.Manager, lm *lock.Manager) *Cleaner {
	return &Cleaner{
		profileManager: pm,
		lockManager:    lm,
	}
}

func (c *Cleaner) Clean(profileID string) (int64, error) {
	p, err := c.profileManager.Get(profileID)
	if err != nil {
		return 0, err
	}

	profileDir := c.profileManager.ProfileDir(p.ID)

	if c.lockManager.IsLocked(profileDir) {
		return 0, fmt.Errorf("cannot clean cache while profile is running: %s", p.Name)
	}

	var totalFreed int64

	for _, dir := range targetDirs {
		dirPath := filepath.Join(profileDir, dir)

		size, err := dirSize(dirPath)
		if err != nil {
			continue
		}

		if err := os.RemoveAll(dirPath); err != nil {
			if logger.Error != nil {
				logger.Error.Printf("failed to remove cache dir %s: %v", dirPath, err)
			}
			continue
		}

		totalFreed += size
	}

	if logger.Info != nil {
		logger.Info.Printf("cleaned cache for profile %s, freed %d bytes", p.Name, totalFreed)
	}

	return totalFreed, nil
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

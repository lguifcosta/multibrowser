package cache

import (
	"os"
	"path/filepath"
	"testing"

	"multibrowser/internal/lock"
	"multibrowser/internal/profile"
)

func setupTest(t *testing.T) (*Cleaner, *profile.Manager, *lock.Manager) {
	t.Helper()
	tmpDir := t.TempDir()

	pm, err := profile.NewManager(tmpDir)
	if err != nil {
		t.Fatalf("failed to create profile manager: %v", err)
	}

	lm := lock.NewManager()
	c := NewCleaner(pm, lm)
	return c, pm, lm
}

func TestClean(t *testing.T) {
	cleaner, pm, _ := setupTest(t)

	p, _ := pm.Create("Test Profile")
	profileDir := pm.ProfileDir(p.ID)

	// Create cache directories with data
	for _, dir := range targetDirs {
		dirPath := filepath.Join(profileDir, dir)
		os.MkdirAll(dirPath, 0755)
		os.WriteFile(filepath.Join(dirPath, "data.bin"), []byte("cache data here"), 0644)
	}

	// Create a non-cache file that should survive
	os.WriteFile(filepath.Join(profileDir, "important.txt"), []byte("keep me"), 0644)

	freed, err := cleaner.Clean(p.ID)
	if err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	if freed <= 0 {
		t.Error("expected some bytes to be freed")
	}

	// Verify cache dirs removed
	for _, dir := range targetDirs {
		dirPath := filepath.Join(profileDir, dir)
		if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
			t.Errorf("cache directory %s should have been removed", dir)
		}
	}

	// Verify non-cache file survives
	if _, err := os.Stat(filepath.Join(profileDir, "important.txt")); os.IsNotExist(err) {
		t.Error("non-cache file should not be removed")
	}
}

func TestClean_RunningProfile(t *testing.T) {
	cleaner, pm, lm := setupTest(t)

	p, _ := pm.Create("Running Profile")

	// Lock the profile
	lm.Acquire(pm.ProfileDir(p.ID), os.Getpid())

	_, err := cleaner.Clean(p.ID)
	if err == nil {
		t.Error("cleaning cache of running profile should fail")
	}
}

func TestClean_NoCacheDirs(t *testing.T) {
	cleaner, pm, _ := setupTest(t)

	p, _ := pm.Create("Empty Profile")

	freed, err := cleaner.Clean(p.ID)
	if err != nil {
		t.Fatalf("clean should not error on empty profile: %v", err)
	}

	if freed != 0 {
		t.Errorf("expected 0 bytes freed, got %d", freed)
	}
}

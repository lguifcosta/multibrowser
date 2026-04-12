package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestManager(t *testing.T) *Manager {
	t.Helper()
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	return m
}

func TestNewManager_CreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profilesDir := filepath.Join(tmpDir, "profiles")
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		t.Error("profiles directory was not created")
	}

	backupsDir := filepath.Join(tmpDir, "backups")
	if _, err := os.Stat(backupsDir); os.IsNotExist(err) {
		t.Error("backups directory was not created")
	}

	metadataPath := filepath.Join(tmpDir, "metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Error("metadata.json was not created")
	}
}

func TestCreate(t *testing.T) {
	m := setupTestManager(t)

	p, err := m.Create("Test Profile")
	if err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}

	if p.Name != "Test Profile" {
		t.Errorf("expected name 'Test Profile', got '%s'", p.Name)
	}

	if p.ID == "" {
		t.Error("profile ID should not be empty")
	}

	if p.Status != StatusStopped {
		t.Errorf("expected status 'stopped', got '%s'", p.Status)
	}

	// Verify directory was created
	profileDir := m.ProfileDir(p.ID)
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		t.Error("profile directory was not created")
	}
}

func TestList(t *testing.T) {
	m := setupTestManager(t)

	m.Create("Profile 1")
	m.Create("Profile 2")

	profiles, err := m.List()
	if err != nil {
		t.Fatalf("failed to list profiles: %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(profiles))
	}
}

func TestGet(t *testing.T) {
	m := setupTestManager(t)

	created, _ := m.Create("Test Profile")

	got, err := m.Get(created.ID)
	if err != nil {
		t.Fatalf("failed to get profile: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("expected ID '%s', got '%s'", created.ID, got.ID)
	}
}

func TestGet_NotFound(t *testing.T) {
	m := setupTestManager(t)

	_, err := m.Get("nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestDelete(t *testing.T) {
	m := setupTestManager(t)

	p, _ := m.Create("To Delete")

	err := m.Delete(p.ID)
	if err != nil {
		t.Fatalf("failed to delete profile: %v", err)
	}

	// Verify removed from metadata
	profiles, _ := m.List()
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}

	// Verify directory removed
	profileDir := m.ProfileDir(p.ID)
	if _, err := os.Stat(profileDir); !os.IsNotExist(err) {
		t.Error("profile directory should have been removed")
	}
}

func TestDelete_RunningProfile(t *testing.T) {
	m := setupTestManager(t)

	p, _ := m.Create("Running")
	m.UpdateStatus(p.ID, StatusRunning)

	err := m.Delete(p.ID)
	if err == nil {
		t.Error("expected error when deleting running profile")
	}
}

func TestRename(t *testing.T) {
	m := setupTestManager(t)

	p, _ := m.Create("Original")

	err := m.Rename(p.ID, "Renamed")
	if err != nil {
		t.Fatalf("failed to rename: %v", err)
	}

	got, _ := m.Get(p.ID)
	if got.Name != "Renamed" {
		t.Errorf("expected name 'Renamed', got '%s'", got.Name)
	}
}

func TestClone(t *testing.T) {
	m := setupTestManager(t)

	p, _ := m.Create("Original")

	// Create a test file in the profile directory
	testFile := filepath.Join(m.ProfileDir(p.ID), "testfile.txt")
	os.WriteFile(testFile, []byte("test data"), 0644)

	cloned, err := m.Clone(p.ID, "Cloned")
	if err != nil {
		t.Fatalf("failed to clone: %v", err)
	}

	if cloned.ID == p.ID {
		t.Error("cloned profile should have a different ID")
	}

	if cloned.Name != "Cloned" {
		t.Errorf("expected name 'Cloned', got '%s'", cloned.Name)
	}

	// Verify file was copied
	clonedFile := filepath.Join(m.ProfileDir(cloned.ID), "testfile.txt")
	data, err := os.ReadFile(clonedFile)
	if err != nil {
		t.Fatalf("cloned file not found: %v", err)
	}
	if string(data) != "test data" {
		t.Errorf("cloned file content mismatch")
	}
}

func TestClone_SkipsLockFile(t *testing.T) {
	m := setupTestManager(t)

	p, _ := m.Create("Original")

	// Create a .lock file in the profile directory
	lockFile := filepath.Join(m.ProfileDir(p.ID), ".lock")
	os.WriteFile(lockFile, []byte("12345"), 0644)

	cloned, err := m.Clone(p.ID, "Cloned")
	if err != nil {
		t.Fatalf("failed to clone: %v", err)
	}

	// Verify .lock was NOT copied
	clonedLock := filepath.Join(m.ProfileDir(cloned.ID), ".lock")
	if _, err := os.Stat(clonedLock); !os.IsNotExist(err) {
		t.Error(".lock file should not be cloned")
	}
}

func TestClone_RunningProfile(t *testing.T) {
	m := setupTestManager(t)

	p, _ := m.Create("Running")
	m.UpdateStatus(p.ID, StatusRunning)

	_, err := m.Clone(p.ID, "Clone")
	if err == nil {
		t.Error("expected error when cloning running profile")
	}
}

func TestUpdateStatus(t *testing.T) {
	m := setupTestManager(t)

	p, _ := m.Create("Test")

	err := m.UpdateStatus(p.ID, StatusRunning)
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	got, _ := m.Get(p.ID)
	if got.Status != StatusRunning {
		t.Errorf("expected status 'running', got '%s'", got.Status)
	}
}

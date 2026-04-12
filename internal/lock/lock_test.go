package lock

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestAcquireAndRelease(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	pid := os.Getpid()

	err := m.Acquire(tmpDir, pid)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	// Verify lock file exists with correct PID
	data, err := os.ReadFile(filepath.Join(tmpDir, ".lock"))
	if err != nil {
		t.Fatalf("lock file not found: %v", err)
	}

	if string(data) != strconv.Itoa(pid) {
		t.Errorf("expected PID %d, got '%s'", pid, string(data))
	}

	// Release
	err = m.Release(tmpDir)
	if err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}

	// Verify lock file removed
	if _, err := os.Stat(filepath.Join(tmpDir, ".lock")); !os.IsNotExist(err) {
		t.Error("lock file should have been removed")
	}
}

func TestAcquire_AlreadyLocked(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	pid := os.Getpid()
	m.Acquire(tmpDir, pid)

	err := m.Acquire(tmpDir, pid)
	if err == nil {
		t.Error("expected error when acquiring already locked profile")
	}
}

func TestIsLocked_ActiveProcess(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	pid := os.Getpid() // Our own PID, which is alive
	m.Acquire(tmpDir, pid)

	if !m.IsLocked(tmpDir) {
		t.Error("expected profile to be locked")
	}
}

func TestIsLocked_OrphanedLock(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	// Write a lock with a PID that doesn't exist
	lockFile := filepath.Join(tmpDir, ".lock")
	os.WriteFile(lockFile, []byte("9999999"), 0644)

	if m.IsLocked(tmpDir) {
		t.Error("orphaned lock should be cleaned and return false")
	}

	// Verify lock file was cleaned up
	if _, err := os.Stat(lockFile); !os.IsNotExist(err) {
		t.Error("orphaned lock file should have been removed")
	}
}

func TestIsLocked_CorruptedLockFile(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	lockFile := filepath.Join(tmpDir, ".lock")
	os.WriteFile(lockFile, []byte("not-a-pid"), 0644)

	if m.IsLocked(tmpDir) {
		t.Error("corrupted lock should be cleaned and return false")
	}

	// Verify lock file was cleaned up
	if _, err := os.Stat(lockFile); !os.IsNotExist(err) {
		t.Error("corrupted lock file should have been removed")
	}
}

func TestIsLocked_NoLockFile(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	if m.IsLocked(tmpDir) {
		t.Error("should not be locked without lock file")
	}
}

func TestGetPID(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	pid := os.Getpid()
	m.Acquire(tmpDir, pid)

	got, err := m.GetPID(tmpDir)
	if err != nil {
		t.Fatalf("failed to get PID: %v", err)
	}

	if got != pid {
		t.Errorf("expected PID %d, got %d", pid, got)
	}
}

func TestRelease_NoLockFile(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	// Should not error
	err := m.Release(tmpDir)
	if err != nil {
		t.Errorf("release without lock file should not error: %v", err)
	}
}

package backup

import (
	"os"
	"path/filepath"
	"testing"

	"multibrowser/internal/lock"
	"multibrowser/internal/profile"
)

func setupTest(t *testing.T) (*Service, *profile.Manager) {
	t.Helper()
	tmpDir := t.TempDir()

	pm, err := profile.NewManager(tmpDir)
	if err != nil {
		t.Fatalf("failed to create profile manager: %v", err)
	}

	lm := lock.NewManager()
	svc := NewService(pm, lm)
	return svc, pm
}

func createTestProfile(t *testing.T, pm *profile.Manager) *profile.Profile {
	t.Helper()
	p, err := pm.Create("Test Profile")
	if err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}

	// Create some test data
	testFile := filepath.Join(pm.ProfileDir(p.ID), "test.txt")
	if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	return p
}

func TestExportImport_Plain(t *testing.T) {
	svc, pm := setupTest(t)
	p := createTestProfile(t, pm)

	// Export
	backupPath, err := svc.Export(p.ID, false, "")
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("backup file was not created")
	}

	// Import
	imported, err := svc.Import(backupPath, "Imported Profile", "")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	if imported.ID == p.ID {
		t.Error("imported profile should have new ID")
	}

	// Verify imported data
	importedFile := filepath.Join(pm.ProfileDir(imported.ID), "test.txt")
	data, err := os.ReadFile(importedFile)
	if err != nil {
		t.Fatalf("imported file not found: %v", err)
	}

	if string(data) != "test data" {
		t.Errorf("expected 'test data', got '%s'", string(data))
	}
}

func TestExportImport_Encrypted(t *testing.T) {
	svc, pm := setupTest(t)
	p := createTestProfile(t, pm)

	password := "test-password-123"

	// Export encrypted
	backupPath, err := svc.Export(p.ID, false, password)
	if err != nil {
		t.Fatalf("encrypted export failed: %v", err)
	}

	// Import with correct password
	imported, err := svc.Import(backupPath, "Imported Encrypted", password)
	if err != nil {
		t.Fatalf("encrypted import failed: %v", err)
	}

	// Verify data
	importedFile := filepath.Join(pm.ProfileDir(imported.ID), "test.txt")
	data, err := os.ReadFile(importedFile)
	if err != nil {
		t.Fatalf("imported file not found: %v", err)
	}

	if string(data) != "test data" {
		t.Errorf("expected 'test data', got '%s'", string(data))
	}
}

func TestImport_WrongPassword(t *testing.T) {
	svc, pm := setupTest(t)
	p := createTestProfile(t, pm)

	backupPath, _ := svc.Export(p.ID, false, "correct-password")

	_, err := svc.Import(backupPath, "Should Fail", "wrong-password")
	if err == nil {
		t.Error("import with wrong password should fail")
	}
}

func TestExport_RunningProfile(t *testing.T) {
	svc, pm := setupTest(t)
	p := createTestProfile(t, pm)

	// Simulate running profile by creating lock with our PID
	lm := lock.NewManager()
	lm.Acquire(pm.ProfileDir(p.ID), os.Getpid())

	_, err := svc.Export(p.ID, false, "")
	if err == nil {
		t.Error("export of running profile should fail")
	}
}

func TestExport_ExcludeCache(t *testing.T) {
	svc, pm := setupTest(t)
	p := createTestProfile(t, pm)

	// Create cache directories
	profileDir := pm.ProfileDir(p.ID)
	cacheDirs := []string{"Cache", "Code Cache", "GPUCache", "Crashpad"}
	for _, dir := range cacheDirs {
		dirPath := filepath.Join(profileDir, dir)
		os.MkdirAll(dirPath, 0755)
		os.WriteFile(filepath.Join(dirPath, "data.bin"), []byte("cache"), 0644)
	}

	// Export excluding cache
	backupPath, err := svc.Export(p.ID, true, "")
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	// Import and check cache dirs are not present
	imported, err := svc.Import(backupPath, "No Cache", "")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	importedDir := pm.ProfileDir(imported.ID)
	for _, dir := range cacheDirs {
		dirPath := filepath.Join(importedDir, dir)
		if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
			t.Errorf("cache directory %s should not exist in import", dir)
		}
	}
}

func TestImport_NonexistentFile(t *testing.T) {
	svc, _ := setupTest(t)

	_, err := svc.Import("/nonexistent/backup.tar.gz", "Test", "")
	if err == nil {
		t.Error("import of nonexistent file should fail")
	}
}

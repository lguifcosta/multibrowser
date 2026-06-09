package config

import (
	"os"
	"testing"
)

func TestConfigManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	m, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Test default config
	cfg := m.Get()
	if cfg.BrowserName != "" {
		t.Errorf("expected empty browser name, got %s", cfg.BrowserName)
	}

	// Test SetBrowser
	err = m.SetBrowser("Chrome", "/usr/bin/chrome")
	if err != nil {
		t.Fatalf("failed to set browser: %v", err)
	}

	cfg = m.Get()
	if cfg.BrowserName != "Chrome" || cfg.BrowserPath != "/usr/bin/chrome" {
		t.Errorf("config mismatch after SetBrowser")
	}

	// Test persistence
	m2, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create manager 2: %v", err)
	}
	cfg2 := m2.Get()
	if cfg2.BrowserName != "Chrome" {
		t.Errorf("persistence failed: expected Chrome, got %s", cfg2.BrowserName)
	}

	// Test UpdateConfig
	newCfg := Config{
		BrowserName:       "Firefox",
		BrowserPath:       "/usr/bin/firefox",
		ShowUnassignedTab: true,
	}
	err = m.UpdateConfig(newCfg)
	if err != nil {
		t.Fatalf("failed to update config: %v", err)
	}

	if m.Get().BrowserName != "Firefox" || !m.Get().ShowUnassignedTab {
		t.Errorf("config update failed")
	}
}

func TestNewManagerNonExistentDir(t *testing.T) {
	_, err := NewManager("/non/existent/path/that/should/fail")
	// It shouldn't fail because it only returns error on ReadFile if it's NOT IsNotExist
	// But it might fail later on save.
	if err != nil {
		t.Errorf("did not expect error on initialization even with non-existent dir")
	}
}

package process

import (
	"os"
	"testing"

	"multibrowser/internal/config"
	"multibrowser/internal/lock"
	"multibrowser/internal/profile"
)

func TestManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "process_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	pm, err := profile.NewManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create profile manager: %v", err)
	}

	lm := lock.NewManager()

	cm, err := config.NewManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}

	m := NewManager(pm, lm, cm)

	// Test Get/Set Browser
	err = m.SetBrowser("TestBrowser", "/usr/bin/testbrowser")
	if err != nil {
		t.Fatalf("failed to set browser: %v", err)
	}

	if m.GetBrowserName() != "TestBrowser" {
		t.Errorf("expected TestBrowser, got %s", m.GetBrowserName())
	}
	if m.GetBrowserPath() != "/usr/bin/testbrowser" {
		t.Errorf("expected /usr/bin/testbrowser, got %s", m.GetBrowserPath())
	}
	if !m.HasBrowser() {
		t.Error("expected HasBrowser to be true")
	}

	// Test IsRunning for non-existent profile
	if m.IsRunning("non-existent") {
		t.Error("expected IsRunning to be false for non-existent profile")
	}
}

func TestBuildArgs(t *testing.T) {
	profileDir := "/tmp/profile1"
	flags := profile.ProfileFlags{
		RestoreLastSession: true,
		UserAgent:          "Mozilla/5.0",
		Lang:               "en-US",
		WindowSize:         "1024,768",
		ProxyServer:        "http://proxy:8080",
		DisableExtensions:  true,
	}

	args := buildArgs(profileDir, flags)

	expectedArgs := []string{
		"--user-data-dir=" + profileDir,
		"--restore-last-session",
		"--user-agent=Mozilla/5.0",
		"--lang=en-US",
		"--window-size=1024,768",
		"--proxy-server=http://proxy:8080",
		"--disable-extensions",
	}

	for _, expected := range expectedArgs {
		found := false
		for _, arg := range args {
			if arg == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing expected arg: %s", expected)
		}
	}
}

func TestDetectBrowsers(t *testing.T) {
	browsers := DetectBrowsers()
	// This will depend on the environment, but it shouldn't crash
	t.Logf("Found %d browsers", len(browsers))
}

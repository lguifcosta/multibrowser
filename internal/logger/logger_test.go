package logger

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoggerInit(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	err = Init(tempDir)
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	if Info == nil {
		t.Error("Info logger is nil")
	}
	if Error == nil {
		t.Error("Error logger is nil")
	}

	logFile := filepath.Join(tempDir, "logs", "app.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("log file not created at %s", logFile)
	}

	Info.Println("Test info message")
	Error.Println("Test error message")
}

func TestInitInvalidPath(t *testing.T) {
	// Try to init in a path that cannot be created (e.g. file exists where dir should be)
	tempFile, err := os.CreateTemp("", "logger_fail")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	err = Init(tempFile.Name())
	if err == nil {
		t.Error("expected error when initializing logger in an invalid path")
	}
}

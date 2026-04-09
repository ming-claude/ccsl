package debug

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// resetState resets all package-level state so each test starts clean.
func resetState(t *testing.T) {
	t.Helper()
	logPath = ""
	pathOnce = sync.Once{}
	enabled.Store(false)
}

func TestSetEnabled(t *testing.T) {
	resetState(t)

	if Enabled() {
		t.Error("Enabled() should default to false")
	}

	SetEnabled(true)
	if !Enabled() {
		t.Error("Enabled() should return true after SetEnabled(true)")
	}

	SetEnabled(false)
	if Enabled() {
		t.Error("Enabled() should return false after SetEnabled(false)")
	}
}

func TestLog_WritesEntry(t *testing.T) {
	resetState(t)

	dir := t.TempDir()
	logPath = filepath.Join(dir, "debug.jsonl")
	pathOnce.Do(func() {}) // mark paths as initialized

	Log(&Entry{
		ConfigStyle: "nerd",
		ConfigLines: 2,
		DurationMs:  42,
	})

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)
	if len(content) == 0 {
		t.Fatal("log file should not be empty")
	}
	if !strings.Contains(content, `"config_style":"nerd"`) {
		t.Errorf("log entry missing config_style field, got: %s", content)
	}
	if !strings.Contains(content, `"duration_ms":42`) {
		t.Errorf("log entry missing duration_ms field, got: %s", content)
	}
	if !strings.Contains(content, `"ts":"`) {
		t.Errorf("log entry missing auto-filled timestamp, got: %s", content)
	}
}

func TestPaths_UnderHomeDir(t *testing.T) {
	resetState(t)

	initPaths()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	expectedDir := filepath.Join(home, ".claude", "ccsl")
	if filepath.Dir(logPath) != expectedDir {
		t.Errorf("logPath should be under %s, got %s", expectedDir, logPath)
	}
}

func TestPaths_NotTmp(t *testing.T) {
	resetState(t)
	initPaths()

	if strings.HasPrefix(logPath, "/tmp") {
		t.Errorf("logPath should not be under /tmp, got %s", logPath)
	}
}

func TestLogFile_Permissions(t *testing.T) {
	resetState(t)

	dir := t.TempDir()
	logPath = filepath.Join(dir, "sub", "debug.jsonl")
	pathOnce.Do(func() {}) // mark paths as initialized

	Log(&Entry{ConfigStyle: "test"})

	// Check log file permissions.
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("log file permissions should be 0600, got %04o", perm)
	}

	// Check parent directory permissions.
	dirInfo, err := os.Stat(filepath.Dir(logPath))
	if err != nil {
		t.Fatalf("log directory not created: %v", err)
	}
	dirPerm := dirInfo.Mode().Perm()
	if dirPerm != 0700 {
		t.Errorf("log directory permissions should be 0700, got %04o", dirPerm)
	}
}

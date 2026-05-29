package snapshot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := Path()
	if err != nil {
		t.Fatalf("Path() error: %v", err)
	}
	want := filepath.Join(home, ".claude", FileName)
	if got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"typical payload", []byte(`{"model":{"id":"claude"},"cost":{"total_cost_usd":1.5}}`)},
		{"empty object", []byte(`{}`)},
		{"verbatim non-json bytes", []byte("written as-is")},
		{"empty input", []byte("")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)

			if err := Write(tt.data); err != nil {
				t.Fatalf("Write() error: %v", err)
			}

			path := filepath.Join(home, ".claude", FileName)
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read back snapshot: %v", err)
			}
			if string(got) != string(tt.data) {
				t.Errorf("snapshot content = %q, want %q", got, tt.data)
			}
			// Temp file must not linger after a successful atomic write.
			if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
				t.Errorf("temp file not cleaned up: stat err = %v", err)
			}
		})
	}
}

func TestWriteCreatesClaudeDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// ~/.claude does not exist yet; Write must create it.
	if err := Write([]byte(`{}`)); err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude")); err != nil {
		t.Errorf(".claude dir not created: %v", err)
	}
}

func TestWriteOverwrites(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := Write([]byte(`{"v":1}`)); err != nil {
		t.Fatalf("first Write() error: %v", err)
	}
	if err := Write([]byte(`{"v":2}`)); err != nil {
		t.Fatalf("second Write() error: %v", err)
	}

	path := filepath.Join(home, ".claude", FileName)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back snapshot: %v", err)
	}
	if string(got) != `{"v":2}` {
		t.Errorf("overwrite failed: content = %q, want %q", got, `{"v":2}`)
	}
}

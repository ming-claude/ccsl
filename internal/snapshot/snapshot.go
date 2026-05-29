// Package snapshot mirrors the Claude Code statusline payload to
// ~/.claude/statusline-snapshot.json so external tools (menu-bar apps,
// dashboards) can read the live session state without installing their own
// statusline bridge.
//
// The payload is written verbatim: it is the raw JSON Claude Code pipes to the
// statusline on stdin, so any tool reading Claude Code's standard fields
// (model, cost, context_window, rate_limits, ...) works unchanged. Freshness is
// conveyed by the file's mtime.
package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileName is the fixed filename external tools look for under ~/.claude.
const FileName = "statusline-snapshot.json"

// Path returns the absolute path to the snapshot file
// (~/.claude/statusline-snapshot.json).
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".claude", FileName), nil
}

// Write atomically writes data to the snapshot file, creating ~/.claude if it
// does not exist. data is written byte-for-byte (no re-encoding).
func Write(data []byte) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}
	return atomicWrite(path, data)
}

// atomicWrite writes data via a temp file + fsync + rename for crash safety,
// matching the pattern used elsewhere in the codebase.
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create temp snapshot: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("write temp snapshot: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("sync temp snapshot: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close temp snapshot: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename temp snapshot: %w", err)
	}
	return nil
}

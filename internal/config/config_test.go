package config

import (
	"os"
	"path/filepath"
	"testing"

	gojson "github.com/goccy/go-json"
)

func TestResolve_DefaultsWhenNilUserConfig(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	cfg, err := Resolve(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Default preset/theme/style.
	if cfg.Preset != "standard" {
		t.Errorf("got preset %q, want %q", cfg.Preset, "standard")
	}
	if cfg.Theme != "catppuccin-block" {
		t.Errorf("got theme %q, want %q", cfg.Theme, "catppuccin-block")
	}
	if cfg.Style != "nerd-font" {
		t.Errorf("got style %q, want %q", cfg.Style, "nerd-font")
	}

	// Lines from standard preset.
	if len(cfg.Lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(cfg.Lines))
	}

	// Theme colors should be populated.
	if len(cfg.ThemeColors) == 0 {
		t.Error("expected non-empty ThemeColors")
	}

	// Separator from block theme (empty string separator is valid; SeparatorSet should be true).
	if !cfg.SeparatorSet {
		t.Error("expected SeparatorSet=true from block theme with explicit separator")
	}

	// Disabled should be empty for standard preset with no user overrides.
	if len(cfg.Disabled) != 0 {
		t.Errorf("got %d disabled entries, want 0", len(cfg.Disabled))
	}
}

func TestResolve_ValidUserConfig(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	uc := &UserConfig{
		Preset:           "full",
		Theme:            "catppuccin-block",
		Style:            "nerd-font",
		DisabledSegments: []string{"speed"},
	}

	cfg, err := Resolve(uc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Preset != "full" {
		t.Errorf("got preset %q, want %q", cfg.Preset, "full")
	}
	if cfg.Style != "nerd-font" {
		t.Errorf("got style %q, want %q", cfg.Style, "nerd-font")
	}

	// Full preset has 6 lines.
	if len(cfg.Lines) != 6 {
		t.Errorf("got %d lines, want 6", len(cfg.Lines))
	}

	// Disabled should contain "speed" from user.
	if !cfg.Disabled["speed"] {
		t.Error("expected 'speed' in Disabled")
	}

	// Theme colors populated.
	if len(cfg.ThemeColors) == 0 {
		t.Error("expected non-empty ThemeColors")
	}
}

func TestResolve_SnapshotFlag(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	tests := []struct {
		name string
		uc   *UserConfig
		want bool
	}{
		{"default off", &UserConfig{}, false},
		{"enabled", &UserConfig{Snapshot: true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Resolve(tt.uc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Snapshot != tt.want {
				t.Errorf("cfg.Snapshot = %v, want %v", cfg.Snapshot, tt.want)
			}
		})
	}
}

func TestResolve_PresetDisabledMergesWithUserDisabled(t *testing.T) {
	// Minimal preset has default disabled entries.
	uc := &UserConfig{
		Preset:           "minimal",
		DisabledSegments: []string{"context.length", "extra.nonexistent"},
	}

	cfg, err := Resolve(uc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have all preset disabled entries.
	presetDisabled := []string{"git.changes", "git.insertions", "git.deletions", "git.worktree", "context.length", "context.pct"}
	for _, s := range presetDisabled {
		if !cfg.Disabled[s] {
			t.Errorf("expected %q in Disabled from preset", s)
		}
	}

	// User's extra non-existent segment should be silently included.
	if !cfg.Disabled["extra.nonexistent"] {
		t.Error("expected 'extra.nonexistent' in Disabled from user")
	}

	// context.length appears in both preset and user — should still be true (union).
	if !cfg.Disabled["context.length"] {
		t.Error("expected 'context.length' in Disabled (present in both preset and user)")
	}
}

func TestResolve_DifferentPresets(t *testing.T) {
	tests := []struct {
		preset       string
		wantLines    int
		wantDisabled int // number of disabled entries from preset alone
	}{
		{"minimal", 1, 6},
		{"standard", 3, 0},
		{"full", 6, 0},
		{"dev", 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.preset, func(t *testing.T) {
			uc := &UserConfig{Preset: tt.preset}
			cfg, err := Resolve(uc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(cfg.Lines) != tt.wantLines {
				t.Errorf("got %d lines, want %d", len(cfg.Lines), tt.wantLines)
			}
			if len(cfg.Disabled) != tt.wantDisabled {
				t.Errorf("got %d disabled, want %d", len(cfg.Disabled), tt.wantDisabled)
			}
		})
	}
}

func TestResolve_NonexistentDisabledSegmentSilentlyIgnored(t *testing.T) {
	uc := &UserConfig{
		DisabledSegments: []string{"nonexistent.group", "also.fake"},
	}

	cfg, err := Resolve(uc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Non-existent segments are just passed through into the disabled map.
	if !cfg.Disabled["nonexistent.group"] {
		t.Error("expected 'nonexistent.group' in Disabled")
	}
	if !cfg.Disabled["also.fake"] {
		t.Error("expected 'also.fake' in Disabled")
	}
}

func TestLoadPreset_AllPresets(t *testing.T) {
	for _, name := range []string{"minimal", "standard", "full", "dev"} {
		t.Run(name, func(t *testing.T) {
			pc, err := LoadPreset(name)
			if err != nil {
				t.Fatalf("unexpected error loading preset %q: %v", name, err)
			}
			if len(pc.Lines) == 0 {
				t.Error("lines should not be empty")
			}
			// Verify lines contain only non-empty group names.
			for i, line := range pc.Lines {
				if len(line) == 0 {
					t.Errorf("line %d is empty", i)
				}
				for j, group := range line {
					if group == "" {
						t.Errorf("line %d, element %d is empty string", i, j)
					}
				}
			}
		})
	}
}

func TestLoadPreset_Unknown(t *testing.T) {
	_, err := LoadPreset("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown preset, got nil")
	}
}

func TestLoadTheme_AllThemes(t *testing.T) {
	expectedGroups := []string{
		"model", "git", "context", "tokens", "cost", "usage_5hour", "usage_weekly",
		"version", "session", "speed", "diff", "activity", "live", "cwd", "env", "clock",
	}

	for _, name := range ListThemes() {
		t.Run(name, func(t *testing.T) {
			theme, err := LoadTheme(name)
			if err != nil {
				t.Fatalf("unexpected error loading theme %q: %v", name, err)
			}
			if len(theme.Colors) == 0 {
				t.Error("theme should have colors")
			}

			// Verify all 16 groups are present.
			for _, grp := range expectedGroups {
				if _, ok := theme.Colors[grp]; !ok {
					t.Errorf("theme missing group color %q", grp)
				}
			}

			// Verify exactly 16 color entries (no leftover old segment keys).
			if len(theme.Colors) != 16 {
				t.Errorf("got %d color entries, want 16", len(theme.Colors))
			}

			if theme.Variant == "block" {
				if c, ok := theme.Colors["model"]; ok {
					if c.Bg == "" {
						t.Error("block theme model color should have bg set")
					}
				}
			}

			// Verify groups with semantic roles have positive and negative set.
			for _, grp := range []string{"git", "diff"} {
				c, ok := theme.Colors[grp]
				if !ok {
					continue
				}
				if c.Positive == "" {
					t.Errorf("theme %q group %q should have positive color", name, grp)
				}
				if c.Negative == "" {
					t.Errorf("theme %q group %q should have negative color", name, grp)
				}
			}
		})
	}
}

func TestLoadTheme_Unknown(t *testing.T) {
	_, err := LoadTheme("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown theme, got nil")
	}
}

func TestResolve_FallbackOnMissingTheme(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	uc := &UserConfig{
		Theme: "nonexistent-theme",
	}
	cfg, err := Resolve(uc)
	if err != nil {
		t.Fatalf("expected fallback to default theme, got error: %v", err)
	}
	if cfg.Theme != defaultTheme {
		t.Errorf("got theme %q after fallback, want %q", cfg.Theme, defaultTheme)
	}
	if len(cfg.ThemeColors) == 0 {
		t.Error("expected non-empty ThemeColors after fallback")
	}
}

func TestListThemes(t *testing.T) {
	names := ListThemes()
	if len(names) < 15 {
		t.Errorf("expected at least 15 themes, got %d: %v", len(names), names)
	}
}

func TestSegmentColor_UnmarshalJSON_String(t *testing.T) {
	var c SegmentColor
	if err := c.UnmarshalJSON([]byte(`"#ff0000"`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Fg != "#ff0000" {
		t.Errorf("got fg %q, want %q", c.Fg, "#ff0000")
	}
	if c.Bg != "" {
		t.Errorf("got bg %q, want empty", c.Bg)
	}
}

func TestSegmentColor_UnmarshalJSON_Object(t *testing.T) {
	var c SegmentColor
	if err := c.UnmarshalJSON([]byte(`{"fg": "#282c34", "bg": "#c678dd"}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Fg != "#282c34" {
		t.Errorf("got fg %q, want %q", c.Fg, "#282c34")
	}
	if c.Bg != "#c678dd" {
		t.Errorf("got bg %q, want %q", c.Bg, "#c678dd")
	}
}

func TestSegmentColor_UnmarshalJSON_ObjectWithSemanticRoles(t *testing.T) {
	var c SegmentColor
	if err := c.UnmarshalJSON([]byte(`{"fg": "#98d885", "bg": "#2d4530", "positive": "#98d885", "negative": "#e87070"}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Fg != "#98d885" {
		t.Errorf("got fg %q, want %q", c.Fg, "#98d885")
	}
	if c.Bg != "#2d4530" {
		t.Errorf("got bg %q, want %q", c.Bg, "#2d4530")
	}
	if c.Positive != "#98d885" {
		t.Errorf("got positive %q, want %q", c.Positive, "#98d885")
	}
	if c.Negative != "#e87070" {
		t.Errorf("got negative %q, want %q", c.Negative, "#e87070")
	}
}

func TestSegmentColor_UnmarshalJSON_NonBlockWithSemanticRoles(t *testing.T) {
	// Non-block themes can also have semantic roles (object form without bg).
	var c SegmentColor
	if err := c.UnmarshalJSON([]byte(`{"fg": "#98d885", "positive": "#98d885", "negative": "#e87070"}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Fg != "#98d885" {
		t.Errorf("got fg %q, want %q", c.Fg, "#98d885")
	}
	if c.Bg != "" {
		t.Errorf("got bg %q, want empty", c.Bg)
	}
	if c.Positive != "#98d885" {
		t.Errorf("got positive %q, want %q", c.Positive, "#98d885")
	}
	if c.Negative != "#e87070" {
		t.Errorf("got negative %q, want %q", c.Negative, "#e87070")
	}
}

func TestSegmentColor_MarshalJSON_FgOnly(t *testing.T) {
	c := SegmentColor{Fg: "red"}
	data, err := c.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `"red"` {
		t.Errorf("got %s, want %q", data, `"red"`)
	}
}

func TestSegmentColor_MarshalJSON_WithBg(t *testing.T) {
	c := SegmentColor{Fg: "#282c34", Bg: "#c678dd"}
	data, err := c.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be object form.
	var m map[string]string
	if err := gojson.Unmarshal(data, &m); err != nil {
		t.Fatalf("should be valid JSON object: %v", err)
	}
	if m["fg"] != "#282c34" || m["bg"] != "#c678dd" {
		t.Errorf("got %s, want fg=#282c34 bg=#c678dd", data)
	}
}

func TestSegmentColor_MarshalJSON_WithSemanticRoles(t *testing.T) {
	c := SegmentColor{Fg: "#98d885", Bg: "#2d4530", Positive: "#98d885", Negative: "#e87070"}
	data, err := c.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be object form with all fields.
	var roundtrip SegmentColor
	if err := gojson.Unmarshal(data, &roundtrip); err != nil {
		t.Fatalf("should be valid JSON: %v", err)
	}
	if roundtrip.Fg != "#98d885" || roundtrip.Bg != "#2d4530" {
		t.Errorf("got fg=%q bg=%q, want fg=#98d885 bg=#2d4530", roundtrip.Fg, roundtrip.Bg)
	}
	if roundtrip.Positive != "#98d885" || roundtrip.Negative != "#e87070" {
		t.Errorf("got positive=%q negative=%q, want positive=#98d885 negative=#e87070", roundtrip.Positive, roundtrip.Negative)
	}
}

func TestSegmentColor_MarshalJSON_PositiveNegativeOnlyUsesObjectForm(t *testing.T) {
	// Even without Bg, having Positive should use object form.
	c := SegmentColor{Fg: "#98d885", Positive: "#98d885", Negative: "#e87070"}
	data, err := c.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should NOT be a simple string since semantic roles are set.
	var s string
	if err := gojson.Unmarshal(data, &s); err == nil {
		t.Errorf("expected object form, got string %q", s)
	}
}

func TestSave_WritesUserConfigOnly(t *testing.T) {
	// Use a temp dir to avoid writing to real config location.
	tmpDir := t.TempDir()

	uc := &UserConfig{
		Preset:           "full",
		Theme:            "catppuccin-block",
		Style:            "nerd-font",
		DisabledSegments: []string{"speed", "env"},
	}

	path := filepath.Join(tmpDir, "config.json")
	err := saveToPath(uc, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read back the file.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("unexpected error reading file: %v", err)
	}

	// Parse as raw JSON to check top-level keys.
	var raw map[string]gojson.RawMessage
	if err := gojson.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unexpected error parsing JSON: %v", err)
	}

	// Should have exactly 4 top-level keys.
	expectedKeys := []string{"preset", "theme", "style", "disabled_segments"}
	if len(raw) != len(expectedKeys) {
		t.Errorf("got %d top-level keys, want %d; keys: %v", len(raw), len(expectedKeys), keys(raw))
	}
	for _, k := range expectedKeys {
		if _, ok := raw[k]; !ok {
			t.Errorf("missing expected key %q", k)
		}
	}

	// Verify roundtrip.
	var loaded UserConfig
	if err := gojson.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.Preset != uc.Preset {
		t.Errorf("got preset %q, want %q", loaded.Preset, uc.Preset)
	}
	if loaded.Theme != uc.Theme {
		t.Errorf("got theme %q, want %q", loaded.Theme, uc.Theme)
	}
	if loaded.Style != uc.Style {
		t.Errorf("got style %q, want %q", loaded.Style, uc.Style)
	}
	if len(loaded.DisabledSegments) != 2 {
		t.Errorf("got %d disabled_segments, want 2", len(loaded.DisabledSegments))
	}
}

func TestSaveAndLoad_Roundtrip(t *testing.T) {
	tmpDir := t.TempDir()

	uc := &UserConfig{
		Preset:           "dev",
		Theme:            "gruvbox-block",
		Style:            "nerd-font",
		DisabledSegments: []string{"activity", "diff"},
	}

	path := filepath.Join(tmpDir, "config.json")

	// Save.
	if err := saveToPath(uc, path); err != nil {
		t.Fatalf("save error: %v", err)
	}

	// Load back.
	loaded, err := loadUserConfigFromPath(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	// Verify all fields.
	if loaded.Preset != uc.Preset {
		t.Errorf("preset: got %q, want %q", loaded.Preset, uc.Preset)
	}
	if loaded.Theme != uc.Theme {
		t.Errorf("theme: got %q, want %q", loaded.Theme, uc.Theme)
	}
	if loaded.Style != uc.Style {
		t.Errorf("style: got %q, want %q", loaded.Style, uc.Style)
	}
	if len(loaded.DisabledSegments) != len(uc.DisabledSegments) {
		t.Errorf("disabled_segments length: got %d, want %d", len(loaded.DisabledSegments), len(uc.DisabledSegments))
	}
	for i, s := range uc.DisabledSegments {
		if i < len(loaded.DisabledSegments) && loaded.DisabledSegments[i] != s {
			t.Errorf("disabled_segments[%d]: got %q, want %q", i, loaded.DisabledSegments[i], s)
		}
	}
	// Resolve from loaded UserConfig and verify runtime config.
	cfg, err := Resolve(loaded)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	if cfg.Preset != "dev" {
		t.Errorf("resolved preset: got %q, want %q", cfg.Preset, "dev")
	}
	if !cfg.Disabled["activity"] {
		t.Error("expected 'activity' in Disabled after roundtrip")
	}
	if !cfg.Disabled["diff"] {
		t.Error("expected 'diff' in Disabled after roundtrip")
	}
}

// --- Helper functions for testing ---

// saveToPath writes a UserConfig to a specific path (for testing without touching real config dir).
func saveToPath(uc *UserConfig, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := gojson.MarshalIndent(uc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, path)
}

// loadUserConfigFromPath loads a UserConfig from a specific path (for testing).
func loadUserConfigFromPath(path string) (*UserConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var uc UserConfig
	if err := gojson.Unmarshal(data, &uc); err != nil {
		return nil, err
	}
	return &uc, nil
}

// keys returns the keys of a map for diagnostic output.
func keys(m map[string]gojson.RawMessage) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func TestResolve_DowngradesBlockThemeWhenNoTruecolor(t *testing.T) {
	t.Setenv("TERM", "xterm")
	t.Setenv("COLORTERM", "")
	t.Setenv("NO_COLOR", "")
	t.Setenv("FORCE_COLOR", "")

	uc := &UserConfig{Theme: "catppuccin-block"}
	cfg, err := Resolve(uc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Theme != "catppuccin" {
		t.Errorf("got theme %q, want %q (downgraded)", cfg.Theme, "catppuccin")
	}
	if !cfg.SeparatorSet {
		t.Error("expected SeparatorSet=true from non-block theme")
	}
}

func TestResolve_KeepsBlockThemeWhenTruecolor(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("NO_COLOR", "")
	t.Setenv("FORCE_COLOR", "")

	uc := &UserConfig{Theme: "catppuccin-block"}
	cfg, err := Resolve(uc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Theme != "catppuccin-block" {
		t.Errorf("got theme %q, want %q (kept)", cfg.Theme, "catppuccin-block")
	}
}

func TestResolve_KeepsBlockThemeWhenNoNonBlockVariant(t *testing.T) {
	t.Setenv("TERM", "xterm")
	t.Setenv("COLORTERM", "")
	t.Setenv("NO_COLOR", "")
	t.Setenv("FORCE_COLOR", "")

	uc := &UserConfig{Theme: "solarized-light-block"}
	cfg, err := Resolve(uc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Theme != "solarized-light-block" {
		t.Errorf("got theme %q, want %q (graceful degradation)", cfg.Theme, "solarized-light-block")
	}
}

func TestResolve_NonBlockThemeUnaffected(t *testing.T) {
	t.Setenv("TERM", "xterm")
	t.Setenv("COLORTERM", "")
	t.Setenv("NO_COLOR", "")
	t.Setenv("FORCE_COLOR", "")

	uc := &UserConfig{Theme: "catppuccin"}
	cfg, err := Resolve(uc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Theme != "catppuccin" {
		t.Errorf("got theme %q, want %q (unchanged)", cfg.Theme, "catppuccin")
	}
}

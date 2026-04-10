package tui

import (
	"strings"
	"testing"

	"github.com/ming-claude/ccsl/internal/config"
	"github.com/ming-claude/ccsl/internal/render"
	_ "github.com/ming-claude/ccsl/internal/style" // register icons
)

func buildTestConfig(presetName, themeName string) *config.Config {
	preset, err := config.LoadPreset(presetName)
	if err != nil {
		return nil
	}
	theme, err := config.LoadTheme(themeName)
	if err != nil {
		return nil
	}
	themeColors := make(map[string]config.SegmentColor, len(theme.Colors))
	for k, v := range theme.Colors {
		themeColors[k] = v
	}
	var sep string
	var sepSet bool
	if theme.Separator != nil {
		sep = *theme.Separator
		sepSet = true
	}
	disabled := make(map[string]bool)
	for _, s := range preset.Disabled {
		disabled[s] = true
	}
	return &config.Config{
		Preset:       presetName,
		Theme:        themeName,
		Style:        "nerd-font",
		Lines:        preset.Lines,
		Disabled:     disabled,
		ThemeColors:  themeColors,
		Separator:    sep,
		SeparatorSet: sepSet,
	}
}

func TestBuildPreview_LineCount(t *testing.T) {
	tests := []struct {
		preset    string
		wantLines int
	}{
		{"minimal", 1},
		{"standard", 3},
		{"full", 6},
		{"dev", 5},
	}

	for _, tt := range tests {
		t.Run(tt.preset, func(t *testing.T) {
			cfg := buildTestConfig(tt.preset, "catppuccin-block")
			if cfg == nil {
				t.Fatal("failed to build test config")
			}
			lines := buildPreview(cfg, 120)
			if len(lines) != tt.wantLines {
				t.Errorf("expected %d lines, got %d", tt.wantLines, len(lines))
			}
			for i, line := range lines {
				if line == "" {
					t.Errorf("line %d is empty", i)
				}
			}
		})
	}
}

func TestBuildPreview_ContainsANSI(t *testing.T) {
	cfg := buildTestConfig("standard", "catppuccin-block")
	if cfg == nil {
		t.Fatal("failed to build test config")
	}
	lines := buildPreview(cfg, 120)
	for i, line := range lines {
		if !strings.Contains(line, "\x1b[") {
			t.Errorf("line %d has no ANSI codes: %q", i, line)
		}
	}
}

func TestBuildPreview_DisabledGroup(t *testing.T) {
	cfg := buildTestConfig("standard", "catppuccin-block")
	if cfg == nil {
		t.Fatal("failed to build test config")
	}
	// Disable all groups on line 1 except model.
	cfg.Disabled["cwd"] = true
	cfg.Disabled["git"] = true

	lines := buildPreview(cfg, 120)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	// Line 0 should still have model content.
	stripped := render.StripANSI(lines[0])
	if stripped == "" {
		t.Error("line 0 should have model content")
	}
}

func TestBuildPreview_AllGroupsDisabled(t *testing.T) {
	cfg := buildTestConfig("minimal", "catppuccin-block")
	if cfg == nil {
		t.Fatal("failed to build test config")
	}
	// Disable all groups in the minimal preset.
	for _, line := range cfg.Lines {
		for _, group := range line {
			cfg.Disabled[group] = true
		}
	}
	lines := buildPreview(cfg, 120)
	// RenderAllLines returns one entry per line even if empty.
	for _, line := range lines {
		if render.StripANSI(line) != "" {
			t.Errorf("expected empty line, got %q", render.StripANSI(line))
		}
	}
}

func TestBuildPreview_TokensFormat(t *testing.T) {
	cfg := buildTestConfig("standard", "catppuccin-block")
	if cfg == nil {
		t.Fatal("failed to build test config")
	}

	lines := buildPreview(cfg, 200)

	// Find the line containing tokens data (line 1 in standard: context, tokens, cost).
	if len(lines) < 2 {
		t.Fatal("expected at least 2 lines")
	}
	tokensLine := render.StripANSI(lines[1])

	// Real pipeline uses ↓ ↑ ♻ markers from TokensGroup.Render().
	if !strings.Contains(tokensLine, "↓") {
		t.Errorf("tokens line should contain ↓ marker, got: %s", tokensLine)
	}
	if !strings.Contains(tokensLine, "↑") {
		t.Errorf("tokens line should contain ↑ marker, got: %s", tokensLine)
	}
	if !strings.Contains(tokensLine, "♻") {
		t.Errorf("tokens line should contain ♻ marker, got: %s", tokensLine)
	}
}

func TestBuildPreview_NonBlockTheme(t *testing.T) {
	cfg := buildTestConfig("standard", "catppuccin")
	if cfg == nil {
		t.Fatal("failed to build test config")
	}
	lines := buildPreview(cfg, 120)
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	// Non-block theme uses │ separator.
	for _, line := range lines {
		stripped := render.StripANSI(line)
		if !strings.Contains(stripped, "│") {
			t.Errorf("non-block theme should use │ separator, got: %s", stripped)
		}
	}
}

func TestBuildPreview_EmptyLines(t *testing.T) {
	cfg := &config.Config{
		Style: "nerd-font",
		Lines: nil,
	}
	lines := buildPreview(cfg, 120)
	if lines != nil {
		t.Errorf("expected nil for empty lines, got %d lines", len(lines))
	}
}

package pipeline

import (
	"testing"

	"github.com/ming-claude/ccsl/internal/config"
	"github.com/ming-claude/ccsl/internal/segment"
	_ "github.com/ming-claude/ccsl/internal/style" // register icons
)

func TestBuildElements_StandardPreset(t *testing.T) {
	preset, err := config.LoadPreset("standard")
	if err != nil {
		t.Fatalf("LoadPreset: %v", err)
	}

	theme, err := config.LoadTheme("catppuccin-block")
	if err != nil {
		t.Fatalf("LoadTheme: %v", err)
	}

	registry := segment.DefaultRegistry()
	ctx := &segment.RenderContext{
		Style: segment.StyleNerdFont,
	}

	elements := BuildElements(preset.Lines, registry, ctx, theme.Colors, segment.StyleNerdFont)

	if len(elements) != len(preset.Lines) {
		t.Errorf("expected %d lines of elements, got %d", len(preset.Lines), len(elements))
	}

	// With nil data, most groups produce nil output — just verify no panic
	// and the correct number of lines.
}

func TestBuildElements_DisabledGroup(t *testing.T) {
	lines := [][]string{{"model", "git", "cwd"}}
	registry := segment.DefaultRegistry()
	ctx := &segment.RenderContext{
		Style:    segment.StyleNerdFont,
		Disabled: map[string]bool{"git": true},
	}
	themeColors := map[string]config.SegmentColor{
		"model": {Fg: "#d4a6f5", Bg: "#503368"},
		"git":   {Fg: "#98d885", Bg: "#2d4530"},
		"cwd":   {Fg: "#b0b8c8", Bg: "#3c404d"},
	}

	elements := BuildElements(lines, registry, ctx, themeColors, segment.StyleNerdFont)

	if len(elements) != 1 {
		t.Fatalf("expected 1 line, got %d", len(elements))
	}

	// With nil data sources, segments likely return nil (except maybe cwd
	// depending on implementation). The key assertion is that "git" is not
	// present even if it would have produced output.
	for _, elem := range elements[0] {
		// Check that no element has git's color (rough check).
		if elem.Color == "#98d885" && elem.BgColor == "#2d4530" {
			t.Error("git group should have been skipped but found element with git colors")
		}
	}
}

func TestBuildElements_AllDisabled(t *testing.T) {
	lines := [][]string{{"model", "git"}}
	registry := segment.DefaultRegistry()
	ctx := &segment.RenderContext{
		Style:    segment.StyleNerdFont,
		Disabled: map[string]bool{"model": true, "git": true},
	}

	elements := BuildElements(lines, registry, ctx, nil, segment.StyleNerdFont)

	if len(elements) != 1 {
		t.Fatalf("expected 1 line, got %d", len(elements))
	}
	if len(elements[0]) != 0 {
		t.Errorf("expected 0 elements in line, got %d", len(elements[0]))
	}
}

func TestBuildElements_EmptyLines(t *testing.T) {
	elements := BuildElements(nil, segment.DefaultRegistry(), &segment.RenderContext{}, nil, segment.StyleNerdFont)
	if len(elements) != 0 {
		t.Errorf("expected 0 lines, got %d", len(elements))
	}
}

func TestBuildElements_UnknownGroup(t *testing.T) {
	lines := [][]string{{"nonexistent_group"}}
	registry := segment.DefaultRegistry()
	ctx := &segment.RenderContext{Style: segment.StyleNerdFont}

	elements := BuildElements(lines, registry, ctx, nil, segment.StyleNerdFont)

	if len(elements) != 1 {
		t.Fatalf("expected 1 line, got %d", len(elements))
	}
	if len(elements[0]) != 0 {
		t.Errorf("expected 0 elements for unknown group, got %d", len(elements[0]))
	}
}

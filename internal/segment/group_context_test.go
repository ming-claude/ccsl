package segment

import (
	"strings"
	"testing"

	"github.com/ming-claude/ccsl/internal/input"
)

func TestContextGroup_FullRender(t *testing.T) {
	ctx := &RenderContext{
		Stdin: &input.StdinData{
			ContextWindow: &input.ContextWindow{
				TotalInputTokens:  5000,
				TotalOutputTokens: 1000,
				ContextWindowSize: 200000,
				UsedPercentage:    42,
				CurrentUsage: &input.CurrentUsage{
					InputTokens:              3000,
					OutputTokens:             500,
					CacheCreationInputTokens: 200,
					CacheReadInputTokens:     300,
				},
			},
		},
		Style: StylePlain,
	}

	g := &ContextGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	// Should contain percentage.
	if !strings.Contains(out.Primary, "42%") {
		t.Errorf("expected output to contain '42%%', got %q", out.Primary)
	}
	// Should contain the length fraction (e.g., "4k/200k").
	if !strings.Contains(out.Primary, "/") {
		t.Errorf("expected output to contain '/' for length, got %q", out.Primary)
	}
}

func TestContextGroup_DisablePct(t *testing.T) {
	ctx := &RenderContext{
		Stdin: &input.StdinData{
			ContextWindow: &input.ContextWindow{
				TotalInputTokens:  5000,
				TotalOutputTokens: 1000,
				ContextWindowSize: 200000,
				UsedPercentage:    42,
			},
		},
		Style: StylePlain,
		Disabled: map[string]bool{
			"context.pct": true,
		},
	}

	g := &ContextGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if strings.Contains(out.Primary, "%") {
		t.Errorf("expected output not to contain '%%' when pct disabled, got %q", out.Primary)
	}
}

func TestContextGroup_NilStdin(t *testing.T) {
	ctx := &RenderContext{
		Style: StylePlain,
	}

	g := &ContextGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

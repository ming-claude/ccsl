package segment

import (
	"strings"
	"testing"

	"github.com/ming-claude/ccsl/internal/transcript"
)

func TestTokensGroup_FullRender(t *testing.T) {
	ctx := &RenderContext{
		Transcript: &transcript.TranscriptData{
			UsageStats: transcript.UsageStats{
				InputTokens:              4500,
				OutputTokens:             1200,
				CacheCreationInputTokens: 800,
				CacheReadInputTokens:     12000,
			},
		},
		Style: StylePlain,
	}

	g := &TokensGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	// Should contain up/down arrow markers for input/output tokens.
	if !strings.Contains(out.Primary, "↑") {
		t.Errorf("expected output to contain '↑', got %q", out.Primary)
	}
	if !strings.Contains(out.Primary, "↓") {
		t.Errorf("expected output to contain '↓', got %q", out.Primary)
	}
	// Should contain cache recycle marker.
	if !strings.Contains(out.Primary, "♻") {
		t.Errorf("expected output to contain '♻', got %q", out.Primary)
	}
}

func TestTokensGroup_NilTranscript(t *testing.T) {
	ctx := &RenderContext{
		Style: StylePlain,
	}

	g := &TokensGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

package segment

import (
	"testing"

	"github.com/ming-claude/ccsl/internal/input"
)

func TestOutputSpeed_FromStdin(t *testing.T) {
	seg := &OutputSpeedSegment{}
	ctx := &RenderContext{
		Stdin: &input.StdinData{
			Cost: &input.Cost{
				TotalAPIDurationMs: 10000, // 10 seconds
			},
			ContextWindow: &input.ContextWindow{
				TotalOutputTokens: 5000,
			},
		},
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	// 5000 / 10 = 500 tok/s
	want := "500 tok/s"
	if out.Primary != want {
		t.Errorf("got %q, want %q", out.Primary, want)
	}
}

func TestOutputSpeed_NoData(t *testing.T) {
	seg := &OutputSpeedSegment{}
	ctx := &RenderContext{
		Stdin: &input.StdinData{
			Cost: &input.Cost{
				TotalAPIDurationMs: 0,
			},
			ContextWindow: &input.ContextWindow{
				TotalOutputTokens: 5000,
			},
		},
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output (no duration), got %+v", out)
	}
}

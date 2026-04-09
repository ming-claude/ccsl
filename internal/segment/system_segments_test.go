package segment

import (
	"testing"
)

func TestTerminalWidth(t *testing.T) {
	seg := &TerminalWidthSegment{}
	ctx := &RenderContext{}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	// Should be a number (possibly "80" default in CI)
	if out.Primary == "" {
		t.Error("expected non-empty Primary")
	}
	// Verify it's a valid number string
	for _, c := range out.Primary {
		if c < '0' || c > '9' {
			t.Errorf("expected numeric string, got %q", out.Primary)
			break
		}
	}
}

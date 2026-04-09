package segment

import (
	"strings"
	"testing"

	"github.com/ming-claude/ccsl/internal/input"
)

func TestDiffGroup_FullRender(t *testing.T) {
	ctx := &RenderContext{
		Stdin: &input.StdinData{
			Cost: &input.Cost{
				TotalLinesAdded:   42,
				TotalLinesRemoved: 7,
			},
		},
		Style: StylePlain,
	}

	g := &DiffGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if !strings.Contains(out.Primary, "+42") {
		t.Errorf("expected output to contain '+42', got %q", out.Primary)
	}
	if !strings.Contains(out.Primary, "-7") {
		t.Errorf("expected output to contain '-7', got %q", out.Primary)
	}
}

func TestDiffGroup_NilGitData(t *testing.T) {
	ctx := &RenderContext{
		Style: StylePlain,
	}

	g := &DiffGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

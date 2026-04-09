package segment

import (
	"strings"
	"testing"

	"github.com/ming-claude/ccsl/internal/input"
)

func TestCostGroup_FullRender(t *testing.T) {
	ctx := &RenderContext{
		Stdin: &input.StdinData{
			Cost: &input.Cost{
				TotalCostUSD: 1.23,
			},
		},
		Style: StylePlain,
	}

	g := &CostGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if !strings.Contains(out.Primary, "1.23") {
		t.Errorf("expected output to contain '1.23', got %q", out.Primary)
	}
	if !strings.Contains(out.Primary, "$") {
		t.Errorf("expected output to contain '$', got %q", out.Primary)
	}
}

func TestCostGroup_NilTranscript(t *testing.T) {
	ctx := &RenderContext{
		Style: StylePlain,
	}

	g := &CostGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

package segment

import (
	"testing"

	"github.com/ming-claude/ccsl/internal/input"
	"github.com/ming-claude/ccsl/internal/render"
)

func TestModelSegment(t *testing.T) {
	seg := &ModelSegment{}

	tests := []struct {
		name    string
		stdin   *input.StdinData
		wantNil bool
		want    string
	}{
		{
			name:    "nil stdin",
			stdin:   nil,
			wantNil: true,
		},
		{
			name:    "nil model",
			stdin:   &input.StdinData{},
			wantNil: true,
		},
		{
			name:  "display name",
			stdin: &input.StdinData{Model: &input.Model{ID: "claude-opus-4-20250514", DisplayName: "Claude Opus"}},
			want:  "Claude Opus",
		},
		{
			name:  "fallback to ID",
			stdin: &input.StdinData{Model: &input.Model{ID: "claude-opus-4-20250514"}},
			want:  "claude-opus-4-20250514",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &RenderContext{Stdin: tt.stdin}
			out, err := seg.Render(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if out != nil {
					t.Fatalf("expected nil output, got %+v", out)
				}
				return
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}
			if out.Primary != tt.want {
				t.Errorf("got %q, want %q", out.Primary, tt.want)
			}
		})
	}
}

func TestSessionCostSegment(t *testing.T) {
	seg := &SessionCostSegment{}

	tests := []struct {
		name    string
		cost    float64
		wantNil bool
		want    string
	}{
		{
			name: "zero cost",
			cost: 0,
			want: "$0",
		},
		{
			name: "normal cost",
			cost: 1.23,
			want: "$1.23",
		},
		{
			name: "small cost",
			cost: 0.01,
			want: "$0.01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &RenderContext{
				Stdin: &input.StdinData{Cost: &input.Cost{TotalCostUSD: tt.cost}},
			}
			out, err := seg.Render(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if out != nil {
					t.Fatalf("expected nil output, got %+v", out)
				}
				return
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}
			if out.Primary != tt.want {
				t.Errorf("got %q, want %q", out.Primary, tt.want)
			}
		})
	}
}

func TestSessionClockSegment(t *testing.T) {
	seg := &SessionClockSegment{}

	tests := []struct {
		name string
		ms   int64
		want string
	}{
		{name: "hours and minutes", ms: 3661000, want: "1h1m"},
		{name: "minutes and seconds", ms: 125000, want: "2m5s"},
		{name: "seconds only", ms: 45000, want: "45s"},
		{name: "exact hour", ms: 3600000, want: "1h0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &RenderContext{
				Stdin: &input.StdinData{Cost: &input.Cost{TotalDurationMs: tt.ms}},
			}
			out, err := seg.Render(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}
			if out.Primary != tt.want {
				t.Errorf("got %q, want %q", out.Primary, tt.want)
			}
		})
	}
}

func TestContextBarSegment(t *testing.T) {
	seg := &ContextBarSegment{}

	ctx := &RenderContext{
		Stdin: &input.StdinData{
			ContextWindow: &input.ContextWindow{UsedPercentage: 50},
		},
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Primary == "" {
		t.Error("expected non-empty bar")
	}
	// Hardcoded width 10, 50% -> 5 filled + 5 empty = 10 display columns.
	if w := render.DisplayWidth(out.Primary); w != 10 {
		t.Errorf("expected 10 display width bar, got %d: %q", w, out.Primary)
	}
}

func TestNilGracefulDegradation(t *testing.T) {
	segments := []Segment{
		&ModelSegment{},
		&OutputStyleSegment{},
		&VimModeSegment{},
		&SessionNameSegment{},
		&SessionIDSegment{},
		&SessionClockSegment{},
		&SessionCostSegment{},
		&ContextBarSegment{},
		&ContextLengthSegment{},
		&TokensInputSegment{},
		&TokensOutputSegment{},
		&TokensCachedSegment{},
		&TokensTotalSegment{},
		&LinesAddedSegment{},
		&LinesRemovedSegment{},
		&CWDSegment{},
		&GitWorktreeSegment{},
		&ContextPctSegment{},
	}

	for _, seg := range segments {
		t.Run(seg.Name()+"_nil_stdin", func(t *testing.T) {
			ctx := &RenderContext{Stdin: nil}
			out, err := seg.Render(ctx)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", seg.Name(), err)
			}
			if out != nil {
				t.Errorf("expected nil output for %s with nil stdin, got %+v", seg.Name(), out)
			}
		})
	}
}

func TestCWDSegment(t *testing.T) {
	seg := &CWDSegment{}

	tests := []struct {
		name string
		cwd  string
		want string
	}{
		{
			name: "default depth 2",
			cwd:  "/home/user/project/foo/bar",
			want: "foo/bar",
		},
		{
			name: "short path",
			cwd:  "/home",
			want: "/home",
		},
		{
			name: "3 parts with root slash",
			cwd:  "/foo/bar",
			want: "foo/bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &RenderContext{
				Stdin: &input.StdinData{CWD: tt.cwd},
			}
			out, err := seg.Render(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}
			if out.Primary != tt.want {
				t.Errorf("got %q, want %q", out.Primary, tt.want)
			}
		})
	}
}

func TestTokensFormatting(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  string
	}{
		{name: "hundreds of millions", value: 123456000, want: "123M"},
		{name: "tens of millions", value: 84395000, want: "84.4M"},
		{name: "millions", value: 8439500, want: "8.44M"},
		{name: "exact million", value: 1000000, want: "1M"},
		{name: "hundreds of thousands", value: 123000, want: "123k"},
		{name: "large value", value: 12300, want: "12.3k"},
		{name: "exact thousand", value: 1000, want: "1k"},
		{name: "small value", value: 500, want: "500"},
		{name: "zero", value: 0, want: "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTokens(tt.value)
			if got != tt.want {
				t.Errorf("formatTokens(%d) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestGitWorktreeSegment(t *testing.T) {
	seg := &GitWorktreeSegment{}

	tests := []struct {
		name    string
		stdin   *input.StdinData
		wantNil bool
		want    string
	}{
		{
			name:    "nil worktree",
			stdin:   &input.StdinData{},
			wantNil: true,
		},
		{
			name:    "empty name",
			stdin:   &input.StdinData{Worktree: &input.WorktreeInfo{}},
			wantNil: true,
		},
		{
			name:  "name and branch",
			stdin: &input.StdinData{Worktree: &input.WorktreeInfo{Name: "feature", Branch: "feat/abc"}},
			want:  "feature (feat/abc)",
		},
		{
			name:  "name only",
			stdin: &input.StdinData{Worktree: &input.WorktreeInfo{Name: "feature"}},
			want:  "feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &RenderContext{Stdin: tt.stdin}
			out, err := seg.Render(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if out != nil {
					t.Fatalf("expected nil output, got %+v", out)
				}
				return
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}
			if out.Primary != tt.want {
				t.Errorf("got %q, want %q", out.Primary, tt.want)
			}
		})
	}
}

func TestContextPctSegment(t *testing.T) {
	seg := &ContextPctSegment{}

	tests := []struct {
		name string
		pct  int
		want string
	}{
		{name: "low usage", pct: 3, want: "3%"},
		{name: "mid range", pct: 27, want: "27%"},
		{name: "high usage", pct: 80, want: "80%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &RenderContext{
				Stdin: &input.StdinData{
					ContextWindow: &input.ContextWindow{UsedPercentage: tt.pct},
				},
			}
			out, err := seg.Render(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}
			if out.Primary != tt.want {
				t.Errorf("pct=%d: got %q, want %q", tt.pct, out.Primary, tt.want)
			}
		})
	}
}

func TestSessionIDSegment(t *testing.T) {
	seg := &SessionIDSegment{}

	tests := []struct {
		name    string
		id      string
		wantNil bool
		want    string
	}{
		{name: "empty", id: "", wantNil: true},
		{name: "short id", id: "abc", want: "abc"},
		{name: "long id", id: "abcdef1234567890", want: "abcdef12"},
		{name: "exact 8", id: "12345678", want: "12345678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &RenderContext{Stdin: &input.StdinData{SessionID: tt.id}}
			out, err := seg.Render(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if out != nil {
					t.Fatalf("expected nil, got %+v", out)
				}
				return
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}
			if out.Primary != tt.want {
				t.Errorf("got %q, want %q", out.Primary, tt.want)
			}
		})
	}
}

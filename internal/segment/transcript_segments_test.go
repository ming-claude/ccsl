package segment

import (
	"testing"

	"github.com/ming-claude/ccsl/internal/transcript"
)

func TestToolActive(t *testing.T) {
	seg := &ToolActiveSegment{}
	ctx := &RenderContext{
		Transcript: &transcript.TranscriptData{
			ActiveTools: []transcript.ToolInfo{
				{Name: "Bash", Context: "ls -la /tmp"},
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
	want := "Bash: ls -la /tmp"
	if out.Primary != want {
		t.Errorf("got %q, want %q", out.Primary, want)
	}
}

func TestToolActive_NoActive(t *testing.T) {
	seg := &ToolActiveSegment{}
	ctx := &RenderContext{
		Transcript: &transcript.TranscriptData{
			ActiveTools: []transcript.ToolInfo{},
		},
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

func TestToolActive_SkipsAgentChildren(t *testing.T) {
	seg := &ToolActiveSegment{}
	ctx := &RenderContext{
		Transcript: &transcript.TranscriptData{
			ActiveTools: []transcript.ToolInfo{
				{Name: "Read", Context: "file.go", ParentID: "agent-1"},
				{Name: "Bash", Context: "go test"},
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
	want := "Bash: go test"
	if out.Primary != want {
		t.Errorf("got %q, want %q", out.Primary, want)
	}
}

func TestToolActive_AllAgentChildren(t *testing.T) {
	seg := &ToolActiveSegment{}
	ctx := &RenderContext{
		Transcript: &transcript.TranscriptData{
			ActiveTools: []transcript.ToolInfo{
				{Name: "Read", Context: "file.go", ParentID: "agent-1"},
			},
		},
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestToolStats(t *testing.T) {
	seg := &ToolStatsSegment{}
	ctx := &RenderContext{
		Transcript: &transcript.TranscriptData{
			ToolStats: map[string]int{
				"Read": 5,
				"Edit": 2,
				"Bash": 3,
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
	// Sorted by count desc, then name asc.
	want := "Read \u00d75  Bash \u00d73  Edit \u00d72"
	if out.Primary != want {
		t.Errorf("got %q, want %q", out.Primary, want)
	}
}

func TestToolStats_Empty(t *testing.T) {
	seg := &ToolStatsSegment{}
	ctx := &RenderContext{
		Transcript: &transcript.TranscriptData{
			ToolStats: map[string]int{},
		},
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestToolStats_TieBreak(t *testing.T) {
	seg := &ToolStatsSegment{}
	ctx := &RenderContext{
		Transcript: &transcript.TranscriptData{
			ToolStats: map[string]int{
				"Grep": 2,
				"Glob": 2,
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
	// Same count: sorted alphabetically.
	want := "Glob \u00d72  Grep \u00d72"
	if out.Primary != want {
		t.Errorf("got %q, want %q", out.Primary, want)
	}
}

func TestLiveGroup_AgentOnly(t *testing.T) {
	ctx := &RenderContext{
		Transcript: &transcript.TranscriptData{
			ActiveTools: []transcript.ToolInfo{
				{Name: "Read", Context: "file.go", ParentID: "agent-1"},
			},
			ActiveAgents: []transcript.AgentCallInfo{
				{Description: "Explore codebase", ActiveTool: &transcript.ToolInfo{Name: "Read"}},
			},
		},
		Disabled: map[string]bool{},
	}
	g := &LiveGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	// tool_active returns nil (agent child), only agent_active renders.
	want := "Explore codebase \u2192 Read"
	if out.Primary != want {
		t.Errorf("got %q, want %q", out.Primary, want)
	}
}

func TestLiveGroup_BothSections(t *testing.T) {
	ctx := &RenderContext{
		Transcript: &transcript.TranscriptData{
			ActiveTools: []transcript.ToolInfo{
				{Name: "Bash", Context: "go test"},
			},
			ActiveAgents: []transcript.AgentCallInfo{
				{Description: "Explore codebase", ActiveTool: &transcript.ToolInfo{Name: "Grep"}},
			},
		},
		Disabled: map[string]bool{},
	}
	g := &LiveGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	want := "Bash: go test \u00b7 Explore codebase \u2192 Grep"
	if out.Primary != want {
		t.Errorf("got %q, want %q", out.Primary, want)
	}
	if !out.IsVariable {
		t.Error("expected IsVariable=true")
	}
}

func TestLiveGroup_Empty(t *testing.T) {
	ctx := &RenderContext{
		Transcript: &transcript.TranscriptData{},
		Disabled:   map[string]bool{},
	}
	g := &LiveGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestActivityGroup_StatsOnly(t *testing.T) {
	ctx := &RenderContext{
		Transcript: &transcript.TranscriptData{
			ToolStats: map[string]int{
				"Read": 3,
				"Edit": 1,
			},
			Todos: []transcript.TodoItem{
				{Subject: "Fix bug", Status: "in_progress"},
				{Subject: "Add tests", Status: "pending"},
			},
			TodoCompleted: 1,
		},
		Disabled: map[string]bool{},
	}
	g := &ActivityGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	want := "Read \u00d73  Edit \u00d71  Fix bug (1/2)"
	if out.Primary != want {
		t.Errorf("got %q, want %q", out.Primary, want)
	}
	if !out.IsVariable {
		t.Error("expected IsVariable=true")
	}
}

func TestTokensFromTranscript(t *testing.T) {
	us := transcript.UsageStats{
		InputTokens:              12300,
		OutputTokens:             8100,
		CacheCreationInputTokens: 500,
		CacheReadInputTokens:     1600,
	}
	td := &transcript.TranscriptData{UsageStats: us}
	ctx := &RenderContext{Transcript: td}

	tests := []struct {
		name string
		seg  Segment
		want string
	}{
		{"input", &TokensInputSegment{}, "12.3k"},
		{"output", &TokensOutputSegment{}, "8.1k"},
		{"cached", &TokensCachedSegment{}, "2.1k"},                                // 500+1600
		{"total", &TokensTotalSegment{}, formatTokens(12300 + 8100 + 500 + 1600)}, // 22.5k
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := tt.seg.Render(ctx)
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

func TestTokensFromTranscript_NilTranscript(t *testing.T) {
	ctx := &RenderContext{Transcript: nil}
	segments := []Segment{
		&TokensInputSegment{},
		&TokensOutputSegment{},
		&TokensCachedSegment{},
		&TokensTotalSegment{},
	}
	for _, seg := range segments {
		t.Run(seg.Name(), func(t *testing.T) {
			out, err := seg.Render(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out != nil {
				t.Errorf("expected nil output, got %+v", out)
			}
		})
	}
}

func TestTokensFromTranscript_ZeroUsage(t *testing.T) {
	td := &transcript.TranscriptData{}
	ctx := &RenderContext{Transcript: td}

	out, err := (&TokensInputSegment{}).Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil || out.Primary != "0" {
		t.Errorf("expected '0', got %+v", out)
	}
}

func TestTodoProgress(t *testing.T) {
	seg := &TodoProgressSegment{}
	ctx := &RenderContext{
		Transcript: &transcript.TranscriptData{
			Todos: []transcript.TodoItem{
				{Subject: "Task A", Status: "completed"},
				{Subject: "Task B", Status: "completed"},
				{Subject: "Task C", Status: "in_progress"},
				{Subject: "Task D", Status: "pending"},
				{Subject: "Task E", Status: "pending"},
			},
			TodoCompleted: 3,
		},
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	want := "Task C (3/5)"
	if out.Primary != want {
		t.Errorf("got %q, want %q", out.Primary, want)
	}
}

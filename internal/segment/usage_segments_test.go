package segment

import (
	"strings"
	"testing"
	"time"

	"github.com/ming-claude/ccsl/internal/input"
	"github.com/ming-claude/ccsl/internal/transcript"
)

func TestSessionUsage_FromStdin(t *testing.T) {
	seg := &SessionUsageSegment{}
	ctx := &RenderContext{
		Stdin: &input.StdinData{
			RateLimits: &input.RateLimits{
				FiveHour: &input.RateLimit{UsedPercentage: 42},
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
	// borderless bar (width=10), percentage.
	want := "████▏      42%"
	if out.Primary != want {
		t.Errorf("got %q, want %q", out.Primary, want)
	}
}

func TestSessionUsage_Loading(t *testing.T) {
	seg := &SessionUsageSegment{}
	ctx := &RenderContext{
		Stdin: &input.StdinData{}, // no rate_limits yet
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Primary != "loading..." {
		t.Errorf("got %q, want %q", out.Primary, "loading...")
	}
}

func TestSessionUsage_Outdated(t *testing.T) {
	seg := &SessionUsageSegment{}
	ctx := &RenderContext{
		Stdin: &input.StdinData{
			RateLimits: &input.RateLimits{
				FiveHour: &input.RateLimit{
					UsedPercentage: 30,
					ResetsAt:       time.Now().Add(-1 * time.Hour).Unix(), // already reset
				},
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
	// Stale data: should contain history icon and percentage
	wantIcon := "\U000F02DA"
	if !strings.Contains(out.Primary, wantIcon) {
		t.Errorf("expected history icon in %q", out.Primary)
	}
	if !strings.Contains(out.Primary, "30%") {
		t.Errorf("expected percentage in %q", out.Primary)
	}
}

func TestWeeklyUsage_Outdated(t *testing.T) {
	seg := &WeeklyUsageSegment{}
	ctx := &RenderContext{
		Stdin: &input.StdinData{
			RateLimits: &input.RateLimits{
				SevenDay: &input.RateLimit{
					UsedPercentage: 38,
					ResetsAt:       time.Now().Add(-2 * time.Hour).Unix(),
				},
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
	wantIcon := "\U000F02DA"
	if !strings.Contains(out.Primary, wantIcon) {
		t.Errorf("expected history icon in %q", out.Primary)
	}
	if !strings.Contains(out.Primary, "38%") {
		t.Errorf("expected percentage in %q", out.Primary)
	}
}

func TestSessionUsage_OutdatedByTranscriptAge(t *testing.T) {
	seg := &SessionUsageSegment{}
	ctx := &RenderContext{
		Stdin: &input.StdinData{
			RateLimits: &input.RateLimits{
				FiveHour: &input.RateLimit{
					UsedPercentage: 1,
					ResetsAt:       time.Now().Add(1 * time.Hour).Unix(), // window NOT expired
				},
			},
		},
		Transcript: &transcript.TranscriptData{
			LastResponseTime: time.Now().Add(-15 * time.Minute), // 15 min ago > 10 min threshold
		},
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	wantIcon := "\U000F02DA"
	if !strings.Contains(out.Primary, wantIcon) {
		t.Errorf("expected history icon in %q", out.Primary)
	}
	if !strings.Contains(out.Primary, "1%") {
		t.Errorf("expected percentage in %q", out.Primary)
	}
}

func TestSessionUsage_FreshTranscript(t *testing.T) {
	seg := &SessionUsageSegment{}
	ctx := &RenderContext{
		Stdin: &input.StdinData{
			RateLimits: &input.RateLimits{
				FiveHour: &input.RateLimit{
					UsedPercentage: 50,
					ResetsAt:       time.Now().Add(2 * time.Hour).Unix(),
				},
			},
		},
		Transcript: &transcript.TranscriptData{
			LastResponseTime: time.Now().Add(-3 * time.Minute), // 3 min ago < 10 min threshold
		},
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	// Should show normal percentage without history icon
	if strings.Contains(out.Primary, "\U000F02DA") {
		t.Error("expected no history icon for fresh data")
	}
}

func TestSessionUsage_NoTranscript_TrustsStdin(t *testing.T) {
	seg := &SessionUsageSegment{}
	ctx := &RenderContext{
		Stdin: &input.StdinData{
			RateLimits: &input.RateLimits{
				FiveHour: &input.RateLimit{
					UsedPercentage: 50,
					ResetsAt:       time.Now().Add(2 * time.Hour).Unix(),
				},
			},
		},
		// no Transcript — can't determine freshness
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if strings.Contains(out.Primary, "\U000F02DA") {
		t.Error("expected no history icon without transcript")
	}
}

func TestSessionUsage_NilStdin(t *testing.T) {
	seg := &SessionUsageSegment{}
	ctx := &RenderContext{Stdin: nil}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

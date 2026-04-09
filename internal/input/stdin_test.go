package input

import (
	"strings"
	"testing"
)

func TestReadStdin_ValidJSON(t *testing.T) {
	jsonStr := `{
		"cwd": "/home/user/project",
		"session_id": "sess-123",
		"session_name": "my-session",
		"transcript_path": "/tmp/transcript.json",
		"version": "1.0.0",
		"model": {"id": "claude-opus-4-20250514", "display_name": "Claude Opus"},
		"workspace": {"current_dir": "/home/user/project", "project_dir": "/home/user/project", "added_dirs": ["/extra"]},
		"cost": {"total_cost_usd": 0.05, "total_duration_ms": 1000, "total_api_duration_ms": 800, "total_lines_added": 10, "total_lines_removed": 3},
		"context_window": {
			"total_input_tokens": 5000,
			"total_output_tokens": 1000,
			"context_window_size": 200000,
			"used_percentage": 3,
			"remaining_percentage": 97,
			"current_usage": {"input_tokens": 500, "output_tokens": 100, "cache_creation_input_tokens": 200, "cache_read_input_tokens": 300}
		},
		"rate_limits": {
			"five_hour": {"used_percentage": 10.5, "resets_at": 1700000000},
			"seven_day": {"used_percentage": 5.2, "resets_at": 1700500000}
		},
		"vim": {"mode": "normal"},
		"agent": {"name": "task"},
		"worktree": {"name": "feat-branch", "path": "/tmp/wt", "branch": "feat-branch", "original_cwd": "/home/user/project", "original_branch": "main"},
		"output_style": {"name": "concise"},
		"exceeds_200k_tokens": false
	}`

	got, raw, err := ReadFrom(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("raw bytes should not be empty")
	}

	if got.CWD != "/home/user/project" {
		t.Errorf("CWD = %q, want %q", got.CWD, "/home/user/project")
	}
	if got.SessionID != "sess-123" {
		t.Errorf("SessionID = %q, want %q", got.SessionID, "sess-123")
	}
	if got.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", got.Version, "1.0.0")
	}
	if got.Model == nil {
		t.Fatal("Model is nil")
	}
	if got.Model.ID != "claude-opus-4-20250514" {
		t.Errorf("Model.ID = %q, want %q", got.Model.ID, "claude-opus-4-20250514")
	}
	if got.Workspace == nil {
		t.Fatal("Workspace is nil")
	}
	if len(got.Workspace.AddedDirs) != 1 || got.Workspace.AddedDirs[0] != "/extra" {
		t.Errorf("Workspace.AddedDirs = %v, want [/extra]", got.Workspace.AddedDirs)
	}
	if got.Cost == nil {
		t.Fatal("Cost is nil")
	}
	if got.Cost.TotalCostUSD != 0.05 {
		t.Errorf("Cost.TotalCostUSD = %f, want 0.05", got.Cost.TotalCostUSD)
	}
	if got.ContextWindow == nil {
		t.Fatal("ContextWindow is nil")
	}
	if got.ContextWindow.UsedPercentage != 3 {
		t.Errorf("ContextWindow.UsedPercentage = %d, want 3", got.ContextWindow.UsedPercentage)
	}
	if got.ContextWindow.CurrentUsage == nil {
		t.Fatal("ContextWindow.CurrentUsage is nil")
	}
	if got.RateLimits == nil {
		t.Fatal("RateLimits is nil")
	}
	if got.RateLimits.FiveHour == nil || got.RateLimits.FiveHour.UsedPercentage != 10.5 {
		t.Errorf("RateLimits.FiveHour = %+v, want UsedPercentage=10.5", got.RateLimits.FiveHour)
	}
	if got.Vim == nil || got.Vim.Mode != "normal" {
		t.Errorf("Vim = %+v, want Mode=normal", got.Vim)
	}
	if got.Agent == nil || got.Agent.Name != "task" {
		t.Errorf("Agent = %+v, want Name=task", got.Agent)
	}
	if got.Worktree == nil || got.Worktree.Branch != "feat-branch" {
		t.Errorf("Worktree = %+v, want Branch=feat-branch", got.Worktree)
	}
	if got.OutputStyle == nil || got.OutputStyle.Name != "concise" {
		t.Errorf("OutputStyle = %+v, want Name=concise", got.OutputStyle)
	}
	if got.ExceedsTokens {
		t.Error("ExceedsTokens = true, want false")
	}
}

func TestReadStdin_MissingFields(t *testing.T) {
	jsonStr := `{"cwd": "/tmp", "session_id": "s1"}`

	got, _, err := ReadFrom(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.CWD != "/tmp" {
		t.Errorf("CWD = %q, want %q", got.CWD, "/tmp")
	}
	if got.SessionID != "s1" {
		t.Errorf("SessionID = %q, want %q", got.SessionID, "s1")
	}
	if got.Model != nil {
		t.Errorf("Model = %+v, want nil", got.Model)
	}
	if got.Workspace != nil {
		t.Errorf("Workspace = %+v, want nil", got.Workspace)
	}
	if got.Cost != nil {
		t.Errorf("Cost = %+v, want nil", got.Cost)
	}
	if got.ContextWindow != nil {
		t.Errorf("ContextWindow = %+v, want nil", got.ContextWindow)
	}
	if got.RateLimits != nil {
		t.Errorf("RateLimits = %+v, want nil", got.RateLimits)
	}
	if got.Vim != nil {
		t.Errorf("Vim = %+v, want nil", got.Vim)
	}
	if got.Agent != nil {
		t.Errorf("Agent = %+v, want nil", got.Agent)
	}
	if got.Worktree != nil {
		t.Errorf("Worktree = %+v, want nil", got.Worktree)
	}
	if got.OutputStyle != nil {
		t.Errorf("OutputStyle = %+v, want nil", got.OutputStyle)
	}
}

func TestReadStdin_Empty(t *testing.T) {
	_, _, err := ReadFrom(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

func TestReadStdin_InvalidJSON(t *testing.T) {
	_, _, err := ReadFrom(strings.NewReader("not json at all"))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

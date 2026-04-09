package transcript

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func testdataPath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func TestParseBasic(t *testing.T) {
	data, err := Parse(testdataPath("basic.jsonl"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if data == nil {
		t.Fatal("Parse() returned nil data")
	}

	if len(data.ActiveTools) != 0 {
		t.Errorf("ActiveTools = %d, want 0", len(data.ActiveTools))
	}
	if data.CompletedTools != 2 {
		t.Errorf("CompletedTools = %d, want 2", data.CompletedTools)
	}
	if got := data.ToolStats["Bash"]; got != 1 {
		t.Errorf("ToolStats[Bash] = %d, want 1", got)
	}
	if got := data.ToolStats["Read"]; got != 1 {
		t.Errorf("ToolStats[Read] = %d, want 1", got)
	}

	// LastResponseTime: latest timestamp is "2026-04-06T10:00:05Z" (uuid a3).
	wantTime, _ := time.Parse(time.RFC3339, "2026-04-06T10:00:05Z")
	if !data.LastResponseTime.Equal(wantTime) {
		t.Errorf("LastResponseTime = %v, want %v", data.LastResponseTime, wantTime)
	}

	// UsageStats: 3 assistant messages cumulative.
	// input: 1000+1500+2000=4500, output: 200+400+600=1200
	// cache_creation: 500+200+100=800, cache_read: 3000+4000+5000=12000
	us := data.UsageStats
	if us.InputTokens != 4500 {
		t.Errorf("UsageStats.InputTokens = %d, want 4500", us.InputTokens)
	}
	if us.OutputTokens != 1200 {
		t.Errorf("UsageStats.OutputTokens = %d, want 1200", us.OutputTokens)
	}
	if us.CacheCreationInputTokens != 800 {
		t.Errorf("UsageStats.CacheCreationInputTokens = %d, want 800", us.CacheCreationInputTokens)
	}
	if us.CacheReadInputTokens != 12000 {
		t.Errorf("UsageStats.CacheReadInputTokens = %d, want 12000", us.CacheReadInputTokens)
	}
}

func TestParseAgentProgress(t *testing.T) {
	data, err := Parse(testdataPath("agent_progress.jsonl"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if data == nil {
		t.Fatal("Parse() returned nil data")
	}

	if got := data.ToolStats["WebSearch"]; got != 1 {
		t.Errorf("ToolStats[WebSearch] = %d, want 1", got)
	}
	if data.AgentStats != 1 {
		t.Errorf("AgentStats = %d, want 1", data.AgentStats)
	}
	if len(data.ActiveAgents) != 1 {
		t.Fatalf("ActiveAgents = %d, want 1", len(data.ActiveAgents))
	}
	agent := data.ActiveAgents[0]
	if agent.Description != "research statusline" {
		t.Errorf("agent Description = %q, want %q", agent.Description, "research statusline")
	}
	if agent.SubagentType != "general-purpose" {
		t.Errorf("agent SubagentType = %q, want %q", agent.SubagentType, "general-purpose")
	}

	// UsageStats: 1 assistant + 2 progress messages.
	// input: 500+300+400=1200, output: 100+50+80=230
	// cache_creation: 50+30+20=100, cache_read: 1000+600+800=2400
	us := data.UsageStats
	if us.InputTokens != 1200 {
		t.Errorf("UsageStats.InputTokens = %d, want 1200", us.InputTokens)
	}
	if us.OutputTokens != 230 {
		t.Errorf("UsageStats.OutputTokens = %d, want 230", us.OutputTokens)
	}
	if us.CacheCreationInputTokens != 100 {
		t.Errorf("UsageStats.CacheCreationInputTokens = %d, want 100", us.CacheCreationInputTokens)
	}
	if us.CacheReadInputTokens != 2400 {
		t.Errorf("UsageStats.CacheReadInputTokens = %d, want 2400", us.CacheReadInputTokens)
	}
}

func TestParseMissingFile(t *testing.T) {
	data, err := Parse("/nonexistent/path/transcript.jsonl")
	if err != nil {
		t.Errorf("Parse() error = %v, want nil", err)
	}
	if data != nil {
		t.Errorf("Parse() data = %v, want nil", data)
	}
}

func TestExtractContext(t *testing.T) {
	tests := []struct {
		name  string
		tool  string
		input string
		want  string
	}{
		{
			"bash short",
			"Bash",
			`{"command":"ls -la"}`,
			"ls -la",
		},
		{
			"bash truncated",
			"Bash",
			`{"command":"very-long-command-that-exceeds-thirty-characters-limit"}`,
			"very-long-command-that-exceeds",
		},
		{
			"read file path",
			"Read",
			`{"file_path":"/home/user/project/foo/bar.go"}`,
			"foo/bar.go",
		},
		{
			"grep pattern",
			"Grep",
			`{"pattern":"TODO"}`,
			"TODO",
		},
		{
			"agent description",
			"Agent",
			`{"description":"research topic"}`,
			"research topic",
		},
		{
			"unknown tool",
			"Unknown",
			`{"anything":"value"}`,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractContext(tt.tool, []byte(tt.input))
			if got != tt.want {
				t.Errorf("extractContext(%q, ...) = %q, want %q", tt.tool, got, tt.want)
			}
		})
	}
}

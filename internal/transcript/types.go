package transcript

import "time"

// TranscriptData holds parsed state from a Claude Code transcript JSONL file.
type TranscriptData struct {
	LastResponseTime time.Time `json:"last_response_time"` // most recent API response

	ActiveTools    []ToolInfo     `json:"active_tools"`
	CompletedTools int            `json:"completed_tools"`
	ToolStats      map[string]int `json:"tool_stats"` // tool name → count

	ActiveAgents []AgentCallInfo `json:"active_agents"`
	AgentStats   int             `json:"agent_stats"`

	Todos         []TodoItem `json:"todos"`
	TodoCompleted int        `json:"todo_completed"`

	Skills []string `json:"skills"`

	UsageStats UsageStats `json:"usage_stats"`
}

// UsageStats holds cumulative token usage accumulated from all
// assistant and progress messages in the transcript.
type UsageStats struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// ToolInfo describes a single tool invocation.
type ToolInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Context  string `json:"context,omitempty"`   // human-readable (file path, command snippet)
	ParentID string `json:"parent_id,omitempty"` // parentToolUseID for subagent tools
}

// AgentCallInfo describes an active subagent invocation.
type AgentCallInfo struct {
	ToolUseID    string    `json:"tool_use_id"`
	SubagentType string    `json:"subagent_type,omitempty"`
	Description  string    `json:"description,omitempty"`
	ActiveTool   *ToolInfo `json:"active_tool,omitempty"`
}

// TodoItem represents a single task/todo entry.
type TodoItem struct {
	Subject string `json:"subject"`
	Status  string `json:"status"`
}

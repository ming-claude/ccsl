package input

type StdinData struct {
	CWD            string         `json:"cwd"`
	SessionID      string         `json:"session_id"`
	SessionName    string         `json:"session_name"`
	TranscriptPath string         `json:"transcript_path"`
	Version        string         `json:"version"`
	Model          *Model         `json:"model"`
	Workspace      *Workspace     `json:"workspace"`
	Cost           *Cost          `json:"cost"`
	ContextWindow  *ContextWindow `json:"context_window"`
	RateLimits     *RateLimits    `json:"rate_limits"`
	Vim            *VimInfo       `json:"vim"`
	Agent          *AgentInfo     `json:"agent"`
	Worktree       *WorktreeInfo  `json:"worktree"`
	OutputStyle    *OutputStyle   `json:"output_style"`
	Effort         *EffortInfo    `json:"effort"`
	ExceedsTokens  bool           `json:"exceeds_200k_tokens"`
}

type Model struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type Workspace struct {
	CurrentDir string   `json:"current_dir"`
	ProjectDir string   `json:"project_dir"`
	AddedDirs  []string `json:"added_dirs"`
}

type Cost struct {
	TotalCostUSD       float64 `json:"total_cost_usd"`
	TotalDurationMs    int64   `json:"total_duration_ms"`
	TotalAPIDurationMs int64   `json:"total_api_duration_ms"`
	TotalLinesAdded    int     `json:"total_lines_added"`
	TotalLinesRemoved  int     `json:"total_lines_removed"`
}

type ContextWindow struct {
	TotalInputTokens    int           `json:"total_input_tokens"`
	TotalOutputTokens   int           `json:"total_output_tokens"`
	ContextWindowSize   int           `json:"context_window_size"`
	UsedPercentage      int           `json:"used_percentage"`
	RemainingPercentage int           `json:"remaining_percentage"`
	CurrentUsage        *CurrentUsage `json:"current_usage"`
}

type CurrentUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

type RateLimits struct {
	FiveHour *RateLimit `json:"five_hour"`
	SevenDay *RateLimit `json:"seven_day"`
}

type RateLimit struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
}

type VimInfo struct {
	Mode string `json:"mode"`
}

type AgentInfo struct {
	Name string `json:"name"`
}

type WorktreeInfo struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	Branch         string `json:"branch"`
	OriginalCWD    string `json:"original_cwd"`
	OriginalBranch string `json:"original_branch"`
}

type OutputStyle struct {
	Name string `json:"name"`
}

// EffortInfo carries the current reasoning effort level. Absent when the model
// does not support the effort parameter. Level is one of low/medium/high/xhigh/max.
type EffortInfo struct {
	Level string `json:"level"`
}

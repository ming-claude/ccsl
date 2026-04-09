package debug

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	gojson "github.com/goccy/go-json"

	gitpkg "github.com/ming-claude/ccsl/internal/git"
	"github.com/ming-claude/ccsl/internal/npm"
	"github.com/ming-claude/ccsl/internal/render"
)

// logPath is resolved lazily from the user's home directory.
// Tests can override it before calling Log().
var (
	logPath  string
	pathOnce sync.Once

	enabled atomic.Bool
)

func initPaths() {
	pathOnce.Do(func() {
		home, err := os.UserHomeDir()
		if err != nil {
			// Fallback: use a dot-dir in the current working directory.
			home = "."
		}
		dir := filepath.Join(home, ".claude", "ccsl")
		logPath = filepath.Join(dir, "debug.jsonl")
	})
}

// SetEnabled sets the debug mode flag. Typically called once
// after loading config, before any concurrent access.
func SetEnabled(on bool) {
	enabled.Store(on)
}

// Enabled returns true if debug mode is active.
func Enabled() bool {
	return enabled.Load()
}

// Timing holds per-stage duration in milliseconds.
type Timing struct {
	StdinMs     int64            `json:"stdin_ms"`
	ConfigMs    int64            `json:"config_ms"`
	FetchMs     int64            `json:"fetch_ms"`
	FetchDetail map[string]int64 `json:"fetch_detail,omitempty"`
	RenderMs    int64            `json:"render_ms"`
}

// Entry is a single debug log record written to the JSONL file.
type Entry struct {
	Timestamp     string            `json:"ts"`
	TermWidthRaw  int               `json:"term_width_raw"`
	TermWidthUsed int               `json:"term_width_used"`
	RightReserved int               `json:"right_reserved"`
	ConfigStyle   string            `json:"config_style"`
	ConfigLines   int               `json:"config_lines"`
	Disabled      map[string]bool   `json:"disabled,omitempty"`
	Stdin         gojson.RawMessage `json:"stdin"`

	// Fetched data — structured snapshots of each data source.
	Git        *gitpkg.GitData    `json:"git,omitempty"`
	Transcript *TranscriptSummary `json:"transcript,omitempty"`
	NpmVersion *npm.VersionInfo   `json:"npm_version,omitempty"`
	MemoryPct  string             `json:"memory_pct,omitempty"`

	SmartAlign  *render.AlignDiagram `json:"smart_align,omitempty"`
	OutputLines []string             `json:"output_lines"`
	Timing      *Timing              `json:"timing"`
	DurationMs  int64                `json:"duration_ms"`
}

// TranscriptSummary is a compact view of transcript data for debug logging.
// Only counts and stats are included; full tool/agent details are omitted.
type TranscriptSummary struct {
	LastResponseTime    string         `json:"last_response_time,omitempty"`
	ActiveTools         int            `json:"active_tools"`
	CompletedTools      int            `json:"completed_tools"`
	ToolStats           map[string]int `json:"tool_stats,omitempty"`
	ActiveAgents        int            `json:"active_agents"`
	AgentStats          int            `json:"agent_stats"`
	Todos               int            `json:"todos"`
	TodoCompleted       int            `json:"todo_completed"`
	Skills              []string       `json:"skills,omitempty"`
	InputTokens         int            `json:"input_tokens"`
	OutputTokens        int            `json:"output_tokens"`
	CacheCreationTokens int            `json:"cache_creation_input_tokens"`
	CacheReadTokens     int            `json:"cache_read_input_tokens"`
}

// Log writes a debug entry to the JSONL log file.
// Callers must guard with Enabled() before calling — this function
// does not re-check the flag for performance.
func Log(e *Entry) {
	initPaths()
	if e.Timestamp == "" {
		e.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	data, err := gojson.Marshal(e)
	if err != nil {
		return
	}
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.Write(data)
	_, _ = f.WriteString("\n")
}

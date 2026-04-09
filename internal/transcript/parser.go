package transcript

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	gojson "github.com/goccy/go-json"
)

// maxScanBuf is the maximum line length the scanner will handle (1 MB).
const maxScanBuf = 1 << 20

// Parse reads a transcript JSONL file and returns parsed tool/agent/todo/skill
// data. If the file does not exist, it returns (nil, nil).
func Parse(path string) (data *TranscriptData, retErr error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer func() {
		if cErr := f.Close(); cErr != nil && retErr == nil {
			retErr = cErr
		}
	}()

	data = &TranscriptData{
		ToolStats: make(map[string]int),
	}

	ps := &parserState{
		data:       data,
		pending:    make(map[string]ToolInfo),
		agentTools: make(map[string]ToolInfo),
		agents:     make(map[string]*AgentCallInfo),
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, maxScanBuf), maxScanBuf)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var event rawEvent
		if err := gojson.Unmarshal(line, &event); err != nil {
			continue // skip malformed lines
		}

		switch event.Type {
		case "assistant":
			trackResponseTime(data, event.Timestamp)
			accumulateUsage(&data.UsageStats, event.Message.Usage)
			ps.processContentBlocks(event.Message.Content, "")
		case "progress":
			var pd progressData
			if err := gojson.Unmarshal(event.Data, &pd); err != nil {
				continue
			}
			if pd.Type != "agent_progress" {
				continue
			}
			trackResponseTime(data, event.Timestamp)
			accumulateUsage(&data.UsageStats, pd.Message.Message.Usage)
			ps.processContentBlocks(pd.Message.Message.Content, event.ParentToolUseID)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Build final active tools list in reverse chronological order
	// (newest first) using the insertion-order slice.
	for i := len(ps.pendingOrder) - 1; i >= 0; i-- {
		id := ps.pendingOrder[i]
		if ti, ok := ps.pending[id]; ok {
			data.ActiveTools = append(data.ActiveTools, ti)
		}
	}

	// Build active agents list in reverse chronological order.
	for i := len(ps.agentOrder) - 1; i >= 0; i-- {
		id := ps.agentOrder[i]
		ac, ok := ps.agents[id]
		if !ok {
			continue
		}
		if at, ok := ps.agentTools[ac.ToolUseID]; ok {
			ac.ActiveTool = &at
		}
		data.ActiveAgents = append(data.ActiveAgents, *ac)
	}

	return data, nil
}

// --- raw JSON structures ---

type rawEvent struct {
	Type            string            `json:"type"`
	Timestamp       string            `json:"timestamp"`
	ParentToolUseID string            `json:"parentToolUseID"`
	Message         rawMessage        `json:"message"`
	Data            gojson.RawMessage `json:"data"`
}

type rawMessage struct {
	Content []gojson.RawMessage `json:"content"`
	Usage   rawUsage            `json:"usage"`
}

type rawUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

type progressData struct {
	Type    string        `json:"type"`
	Message progressInner `json:"message"`
}

type progressInner struct {
	Message rawMessage `json:"message"`
}

type contentBlock struct {
	Type      string            `json:"type"`
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Input     gojson.RawMessage `json:"input"`
	ToolUseID string            `json:"tool_use_id"`
}

// parserState holds mutable state accumulated while parsing a transcript.
type parserState struct {
	data         *TranscriptData
	pending      map[string]ToolInfo       // tool_use IDs without a result
	pendingOrder []string                  // insertion order of pending tool IDs
	agentTools   map[string]ToolInfo       // latest active tool per agent
	agents       map[string]*AgentCallInfo // active agent invocations
	agentOrder   []string                  // insertion order of agent IDs
}

func (ps *parserState) processContentBlocks(blocks []gojson.RawMessage, parentID string) {
	for _, raw := range blocks {
		var cb contentBlock
		if err := gojson.Unmarshal(raw, &cb); err != nil {
			continue
		}

		switch cb.Type {
		case "tool_use":
			ps.handleToolUse(cb, parentID)
		case "tool_result":
			ps.handleToolResult(cb, parentID)
		}
	}
}

func (ps *parserState) handleToolUse(cb contentBlock, parentID string) {
	ps.data.ToolStats[cb.Name]++

	ctx := extractContext(cb.Name, cb.Input)
	ti := ToolInfo{
		ID:       cb.ID,
		Name:     cb.Name,
		Context:  ctx,
		ParentID: parentID,
	}

	ps.pending[cb.ID] = ti
	ps.pendingOrder = append(ps.pendingOrder, cb.ID)

	if parentID != "" {
		ps.agentTools[parentID] = ti
	}

	if cb.Name == "Agent" || cb.Name == "Task" {
		var inp struct {
			Description  string `json:"description"`
			SubagentType string `json:"subagent_type"`
		}
		_ = gojson.Unmarshal(cb.Input, &inp)

		ps.agents[cb.ID] = &AgentCallInfo{
			ToolUseID:    cb.ID,
			SubagentType: inp.SubagentType,
			Description:  inp.Description,
		}
		ps.agentOrder = append(ps.agentOrder, cb.ID)
		ps.data.AgentStats++
	}

	if cb.Name == "TodoWrite" || cb.Name == "TaskCreate" {
		var inp struct {
			Todos []struct {
				Content string `json:"content"`
				Status  string `json:"status"`
			} `json:"todos"`
		}
		if err := gojson.Unmarshal(cb.Input, &inp); err == nil {
			for _, todo := range inp.Todos {
				ps.data.Todos = append(ps.data.Todos, TodoItem{
					Subject: todo.Content,
					Status:  todo.Status,
				})
				if todo.Status == "completed" {
					ps.data.TodoCompleted++
				}
			}
		}
	}

	if cb.Name == "TaskUpdate" {
		var inp struct {
			Status string `json:"status"`
		}
		if err := gojson.Unmarshal(cb.Input, &inp); err == nil {
			if inp.Status == "completed" {
				ps.data.TodoCompleted++
				if ps.data.TodoCompleted > len(ps.data.Todos) {
					ps.data.TodoCompleted = len(ps.data.Todos)
				}
			}
		}
	}

	if cb.Name == "Skill" {
		var inp struct {
			Skill string `json:"skill"`
		}
		if err := gojson.Unmarshal(cb.Input, &inp); err == nil && inp.Skill != "" {
			ps.data.Skills = append(ps.data.Skills, inp.Skill)
		}
	}
}

func (ps *parserState) handleToolResult(cb contentBlock, parentID string) {
	if _, ok := ps.pending[cb.ToolUseID]; ok {
		delete(ps.pending, cb.ToolUseID)
		ps.data.CompletedTools++

		if parentID != "" {
			if at, ok := ps.agentTools[parentID]; ok && at.ID == cb.ToolUseID {
				delete(ps.agentTools, parentID)
			}
		}
	}
}

// extractContext produces a short human-readable string describing a tool
// invocation.
func extractContext(name string, input gojson.RawMessage) string {
	if len(input) == 0 {
		return ""
	}

	switch name {
	case "Bash":
		var inp struct {
			Command string `json:"command"`
		}
		if gojson.Unmarshal(input, &inp) == nil && inp.Command != "" {
			if len(inp.Command) > 30 {
				return inp.Command[:30]
			}
			return inp.Command
		}

	case "Read", "Write", "Edit":
		var inp struct {
			FilePath string `json:"file_path"`
		}
		if gojson.Unmarshal(input, &inp) == nil && inp.FilePath != "" {
			return lastNPathComponents(inp.FilePath, 2)
		}

	case "Grep", "Glob":
		var inp struct {
			Pattern string `json:"pattern"`
		}
		if gojson.Unmarshal(input, &inp) == nil && inp.Pattern != "" {
			return inp.Pattern
		}

	case "Agent", "Task":
		var inp struct {
			Description string `json:"description"`
		}
		if gojson.Unmarshal(input, &inp) == nil && inp.Description != "" {
			return inp.Description
		}
	}

	return ""
}

// trackResponseTime updates LastResponseTime if the given timestamp is
// more recent. Only called for assistant/progress events (API responses).
func trackResponseTime(data *TranscriptData, ts string) {
	if ts == "" {
		return
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return
	}
	if t.After(data.LastResponseTime) {
		data.LastResponseTime = t
	}
}

// accumulateUsage adds the token counts from a single message's usage
// into the cumulative stats.
func accumulateUsage(stats *UsageStats, u rawUsage) {
	stats.InputTokens += u.InputTokens
	stats.OutputTokens += u.OutputTokens
	stats.CacheCreationInputTokens += u.CacheCreationInputTokens
	stats.CacheReadInputTokens += u.CacheReadInputTokens
}

// lastNPathComponents returns the last n components of a file path.
func lastNPathComponents(p string, n int) string {
	parts := strings.Split(filepath.ToSlash(p), "/")
	if len(parts) <= n {
		return p
	}
	return strings.Join(parts[len(parts)-n:], "/")
}

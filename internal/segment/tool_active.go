package segment

// ToolActiveSegment displays the currently active tool invocation.
type ToolActiveSegment struct{}

func (s *ToolActiveSegment) Name() string         { return "tool_active" }
func (s *ToolActiveSegment) DefaultTitle() string { return "" }
func (s *ToolActiveSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("tool_active", m)
}
func (s *ToolActiveSegment) DefaultPriority() int { return 6 }

func (s *ToolActiveSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	td := ctx.Transcript
	if td == nil || len(td.ActiveTools) == 0 {
		return nil, nil
	}

	// Show only top-level tools (not agent children).
	// Agent-internal tools are displayed by AgentActiveSegment.
	for _, tool := range td.ActiveTools {
		if tool.ParentID != "" {
			continue
		}
		primary := tool.Name
		if tool.Context != "" {
			primary += ": " + tool.Context
		}
		return &SegmentOutput{Primary: primary}, nil
	}

	return nil, nil
}

package segment

// AgentActiveSegment displays the currently active subagent invocation.
type AgentActiveSegment struct{}

func (s *AgentActiveSegment) Name() string         { return "agent_active" }
func (s *AgentActiveSegment) DefaultTitle() string { return "" }
func (s *AgentActiveSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("agent_active", m)
}
func (s *AgentActiveSegment) DefaultPriority() int { return 6 }

func (s *AgentActiveSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	td := ctx.Transcript
	if td == nil || len(td.ActiveAgents) == 0 {
		return nil, nil
	}

	agent := td.ActiveAgents[0]
	primary := agent.Description
	if agent.ActiveTool != nil {
		primary += " → " + agent.ActiveTool.Name
	}

	return &SegmentOutput{Primary: primary}, nil
}

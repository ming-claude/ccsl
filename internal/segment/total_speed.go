package segment

// TotalSpeedSegment displays the session-average combined (input+output) token throughput.
type TotalSpeedSegment struct{}

func (s *TotalSpeedSegment) Name() string         { return "total_speed" }
func (s *TotalSpeedSegment) DefaultTitle() string { return "" }
func (s *TotalSpeedSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("total_speed", m)
}
func (s *TotalSpeedSegment) DefaultPriority() int { return 4 }

func (s *TotalSpeedSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.Cost == nil || stdin.ContextWindow == nil {
		return nil, nil
	}
	if stdin.Cost.TotalAPIDurationMs == 0 {
		return nil, nil
	}
	tokens := float64(stdin.ContextWindow.TotalInputTokens + stdin.ContextWindow.TotalOutputTokens)
	seconds := float64(stdin.Cost.TotalAPIDurationMs) / 1000.0
	speed := tokens / seconds
	return &SegmentOutput{Primary: formatSpeed(speed)}, nil
}

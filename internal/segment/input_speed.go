package segment

// InputSpeedSegment displays the session-average input token throughput.
type InputSpeedSegment struct{}

func (s *InputSpeedSegment) Name() string         { return "input_speed" }
func (s *InputSpeedSegment) DefaultTitle() string { return "" }
func (s *InputSpeedSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("input_speed", m)
}
func (s *InputSpeedSegment) DefaultPriority() int { return 4 }

func (s *InputSpeedSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.Cost == nil || stdin.ContextWindow == nil {
		return nil, nil
	}
	if stdin.Cost.TotalAPIDurationMs == 0 {
		return nil, nil
	}
	tokens := float64(stdin.ContextWindow.TotalInputTokens)
	seconds := float64(stdin.Cost.TotalAPIDurationMs) / 1000.0
	speed := tokens / seconds
	return &SegmentOutput{Primary: formatSpeed(speed)}, nil
}

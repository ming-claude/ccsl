package segment

// ContextLengthSegment displays context usage as "Xk/Yk" (current/total tokens in thousands).
type ContextLengthSegment struct{}

func (s *ContextLengthSegment) Name() string         { return "context_length" }
func (s *ContextLengthSegment) DefaultTitle() string { return "" }
func (s *ContextLengthSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("context_length", m)
}
func (s *ContextLengthSegment) DefaultPriority() int { return 4 }

func (s *ContextLengthSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.ContextWindow == nil {
		return nil, nil
	}
	cw := stdin.ContextWindow
	var current int
	if cu := cw.CurrentUsage; cu != nil {
		current = cu.InputTokens + cu.OutputTokens + cu.CacheCreationInputTokens + cu.CacheReadInputTokens
	} else {
		current = cw.TotalInputTokens + cw.TotalOutputTokens
	}
	total := cw.ContextWindowSize
	return &SegmentOutput{
		Primary: formatTokens(current) + "/" + formatTokens(total),
	}, nil
}

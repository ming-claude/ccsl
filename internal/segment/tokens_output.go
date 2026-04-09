package segment

// TokensOutputSegment displays the cumulative output token count.
type TokensOutputSegment struct{}

func (s *TokensOutputSegment) Name() string         { return "tokens_output" }
func (s *TokensOutputSegment) DefaultTitle() string { return "" }
func (s *TokensOutputSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("tokens_output", m)
}
func (s *TokensOutputSegment) DefaultPriority() int { return 6 }

func (s *TokensOutputSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	td := ctx.Transcript
	if td == nil {
		return nil, nil
	}
	return &SegmentOutput{
		Primary: formatTokens(td.UsageStats.OutputTokens),
	}, nil
}

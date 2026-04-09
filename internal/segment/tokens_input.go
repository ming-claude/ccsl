package segment

// TokensInputSegment displays the cumulative input token count.
type TokensInputSegment struct{}

func (s *TokensInputSegment) Name() string         { return "tokens_input" }
func (s *TokensInputSegment) DefaultTitle() string { return "" }
func (s *TokensInputSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("tokens_input", m)
}
func (s *TokensInputSegment) DefaultPriority() int { return 6 }

func (s *TokensInputSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	td := ctx.Transcript
	if td == nil {
		return nil, nil
	}
	return &SegmentOutput{
		Primary: formatTokens(td.UsageStats.InputTokens),
	}, nil
}

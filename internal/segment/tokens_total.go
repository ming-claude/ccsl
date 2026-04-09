package segment

// TokensTotalSegment displays the cumulative total token count (all usage fields).
type TokensTotalSegment struct{}

func (s *TokensTotalSegment) Name() string         { return "tokens_total" }
func (s *TokensTotalSegment) DefaultTitle() string { return "" }
func (s *TokensTotalSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("tokens_total", m)
}
func (s *TokensTotalSegment) DefaultPriority() int { return 4 }

func (s *TokensTotalSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	td := ctx.Transcript
	if td == nil {
		return nil, nil
	}
	us := td.UsageStats
	total := us.InputTokens + us.OutputTokens + us.CacheCreationInputTokens + us.CacheReadInputTokens
	return &SegmentOutput{Primary: formatTokens(total)}, nil
}

package segment

// TokensCachedSegment displays the cumulative cached token count
// (cache creation + cache read input tokens).
type TokensCachedSegment struct{}

func (s *TokensCachedSegment) Name() string         { return "tokens_cached" }
func (s *TokensCachedSegment) DefaultTitle() string { return "" }
func (s *TokensCachedSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("tokens_cached", m)
}
func (s *TokensCachedSegment) DefaultPriority() int { return 4 }

func (s *TokensCachedSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	td := ctx.Transcript
	if td == nil {
		return nil, nil
	}
	total := td.UsageStats.CacheCreationInputTokens + td.UsageStats.CacheReadInputTokens
	return &SegmentOutput{Primary: formatTokens(total)}, nil
}

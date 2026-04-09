package segment

// UsageGroup wraps the session_usage (5-hour) segment as a standalone group.
type UsageGroup struct{}

func (g *UsageGroup) Name() string         { return "usage_5hour" }
func (g *UsageGroup) DefaultTitle() string { return "" }
func (g *UsageGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("usage_5hour", m)
}
func (g *UsageGroup) DefaultPriority() int { return 7 }

func (g *UsageGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	seg := &SessionUsageSegment{}
	return seg.Render(ctx)
}

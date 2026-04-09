package segment

// WeeklyGroup wraps the weekly_usage (7-day) segment as a standalone group.
type WeeklyGroup struct{}

func (g *WeeklyGroup) Name() string         { return "usage_weekly" }
func (g *WeeklyGroup) DefaultTitle() string { return "" }
func (g *WeeklyGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("usage_weekly", m)
}
func (g *WeeklyGroup) DefaultPriority() int { return 7 }

func (g *WeeklyGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	seg := &WeeklyUsageSegment{}
	return seg.Render(ctx)
}

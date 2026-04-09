package segment

// ClockGroup wraps the clock segment as a standalone group.
type ClockGroup struct{}

func (g *ClockGroup) Name() string         { return "clock" }
func (g *ClockGroup) DefaultTitle() string { return "" }
func (g *ClockGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("clock", m)
}
func (g *ClockGroup) DefaultPriority() int { return 3 }

func (g *ClockGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	seg := &ClockSegment{}
	return seg.Render(ctx)
}

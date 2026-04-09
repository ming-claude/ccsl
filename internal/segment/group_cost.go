package segment

// CostGroup wraps the session_cost segment as a standalone group.
type CostGroup struct{}

func (g *CostGroup) Name() string         { return "cost" }
func (g *CostGroup) DefaultTitle() string { return "" }
func (g *CostGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("cost", m)
}
func (g *CostGroup) DefaultPriority() int { return 6 }

func (g *CostGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	seg := &SessionCostSegment{}
	return seg.Render(ctx)
}

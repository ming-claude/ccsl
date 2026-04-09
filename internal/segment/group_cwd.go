package segment

// CWDGroup wraps the cwd segment as a standalone group.
type CWDGroup struct{}

func (g *CWDGroup) Name() string         { return "cwd" }
func (g *CWDGroup) DefaultTitle() string { return "" }
func (g *CWDGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("cwd", m)
}
func (g *CWDGroup) DefaultPriority() int { return 4 }

func (g *CWDGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	seg := &CWDSegment{}
	return seg.Render(ctx)
}

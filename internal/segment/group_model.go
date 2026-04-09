package segment

// ModelGroup wraps the model segment.
type ModelGroup struct{}

func (g *ModelGroup) Name() string         { return "model" }
func (g *ModelGroup) DefaultTitle() string { return "" }
func (g *ModelGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("model", m)
}
func (g *ModelGroup) DefaultPriority() int { return 7 }

func (g *ModelGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	if ctx.IsDisabled("model", "model") {
		return nil, nil
	}
	seg := &ModelSegment{}
	return seg.Render(ctx)
}

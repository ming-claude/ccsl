package segment

import "strings"

// ContextGroup composes context_bar, context_pct, and context_length segments.
type ContextGroup struct{}

func (g *ContextGroup) Name() string         { return "context" }
func (g *ContextGroup) DefaultTitle() string { return "" }
func (g *ContextGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("context", m)
}
func (g *ContextGroup) DefaultPriority() int { return 7 }

func (g *ContextGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	var parts []string

	// Bar first, then percentage.
	if !ctx.IsDisabled("context", "bar") {
		seg := &ContextBarSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if !ctx.IsDisabled("context", "pct") {
		seg := &ContextPctSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if !ctx.IsDisabled("context", "length") {
		seg := &ContextLengthSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if len(parts) == 0 {
		return nil, nil
	}
	return &SegmentOutput{Primary: strings.Join(parts, " ")}, nil
}

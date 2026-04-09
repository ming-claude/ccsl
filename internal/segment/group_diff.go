package segment

import "strings"

// DiffGroup composes lines_added and lines_removed segments
// formatted as "+N/-N".
type DiffGroup struct{}

func (g *DiffGroup) Name() string         { return "diff" }
func (g *DiffGroup) DefaultTitle() string { return "" }
func (g *DiffGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("diff", m)
}
func (g *DiffGroup) DefaultPriority() int { return 4 }

func (g *DiffGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	var parts []string

	if !ctx.IsDisabled("diff", "added") {
		seg := &LinesAddedSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, "+"+out.Primary)
		}
	}

	if !ctx.IsDisabled("diff", "removed") {
		seg := &LinesRemovedSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, "-"+out.Primary)
		}
	}

	if len(parts) == 0 {
		return nil, nil
	}
	return &SegmentOutput{Primary: strings.Join(parts, "/")}, nil
}

package segment

import "strings"

// SpeedGroup composes input_speed, output_speed, and total_speed segments.
type SpeedGroup struct{}

func (g *SpeedGroup) Name() string         { return "speed" }
func (g *SpeedGroup) DefaultTitle() string { return "" }
func (g *SpeedGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("speed", m)
}
func (g *SpeedGroup) DefaultPriority() int { return 4 }

func (g *SpeedGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	var parts []string

	if !ctx.IsDisabled("speed", "input") {
		seg := &InputSpeedSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if !ctx.IsDisabled("speed", "output") {
		seg := &OutputSpeedSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if !ctx.IsDisabled("speed", "total") {
		seg := &TotalSpeedSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if len(parts) == 0 {
		return nil, nil
	}
	return &SegmentOutput{Primary: strings.Join(parts, " ")}, nil
}

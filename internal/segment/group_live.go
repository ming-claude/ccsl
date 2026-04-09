package segment

import "strings"

// LiveGroup composes tool_active and agent_active segments
// for real-time status of currently executing operations.
type LiveGroup struct{}

func (g *LiveGroup) Name() string         { return "live" }
func (g *LiveGroup) DefaultTitle() string { return "" }
func (g *LiveGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("live", m)
}
func (g *LiveGroup) DefaultPriority() int { return 4 }

func (g *LiveGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	var parts []string

	if !ctx.IsDisabled("live", "tool_active") {
		seg := &ToolActiveSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if !ctx.IsDisabled("live", "agent_active") {
		seg := &AgentActiveSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if len(parts) == 0 {
		return nil, nil
	}
	return &SegmentOutput{
		Primary:    strings.Join(parts, " \u00b7 "),
		IsVariable: true,
		MinWidth:   12,
	}, nil
}

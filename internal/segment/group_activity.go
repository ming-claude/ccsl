package segment

import "strings"

// ActivityGroup composes tool_stats, todo_progress, and skills segments
// for cumulative session statistics.
type ActivityGroup struct{}

func (g *ActivityGroup) Name() string         { return "activity" }
func (g *ActivityGroup) DefaultTitle() string { return "" }
func (g *ActivityGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("activity", m)
}
func (g *ActivityGroup) DefaultPriority() int { return 5 }

func (g *ActivityGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	var parts []string

	if !ctx.IsDisabled("activity", "tool_stats") {
		seg := &ToolStatsSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if !ctx.IsDisabled("activity", "todo_progress") {
		seg := &TodoProgressSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if !ctx.IsDisabled("activity", "skills") {
		seg := &SkillsSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if len(parts) == 0 {
		return nil, nil
	}
	return &SegmentOutput{
		Primary:    strings.Join(parts, "  "),
		IsVariable: true,
		MinWidth:   15,
	}, nil
}

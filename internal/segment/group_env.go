package segment

import "strings"

// EnvGroup composes vim_mode, output_style, terminal_width,
// and memory_usage segments.
type EnvGroup struct{}

func (g *EnvGroup) Name() string         { return "env" }
func (g *EnvGroup) DefaultTitle() string { return "" }
func (g *EnvGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("env", m)
}
func (g *EnvGroup) DefaultPriority() int { return 4 }

func (g *EnvGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	var parts []string

	if !ctx.IsDisabled("env", "vim_mode") {
		seg := &VimModeSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if !ctx.IsDisabled("env", "output_style") {
		seg := &OutputStyleSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if !ctx.IsDisabled("env", "terminal_width") {
		seg := &TerminalWidthSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if !ctx.IsDisabled("env", "memory_usage") {
		seg := &MemoryUsageSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if len(parts) == 0 {
		return nil, nil
	}
	return &SegmentOutput{Primary: strings.Join(parts, " ")}, nil
}

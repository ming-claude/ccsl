package pipeline

import (
	"github.com/ming-claude/ccsl/internal/config"
	"github.com/ming-claude/ccsl/internal/render"
	"github.com/ming-claude/ccsl/internal/segment"
)

// BuildElements converts configured lines of group names into renderable
// elements by looking up each group in the registry, calling Render(), and
// mapping the output to RenderedElement.
//
// Parameters:
//   - lines: group name arrays per line (from preset via Config.Lines)
//   - registry: segment registry containing group implementations
//   - ctx: shared render context with data sources and disabled map
//   - themeColors: color map from theme, keyed by group name
//   - separator: unused here but kept for API symmetry (render layer uses it)
//   - styleMode: icon style mode (plain, nerd-font, powerline)
func BuildElements(
	lines [][]string,
	registry *segment.Registry,
	ctx *segment.RenderContext,
	themeColors map[string]config.SegmentColor,
	styleMode segment.StyleMode,
) [][]render.RenderedElement {
	var allElements [][]render.RenderedElement

	for _, line := range lines {
		var elements []render.RenderedElement

		for _, groupName := range line {
			// Skip group-level disabled.
			if ctx.Disabled != nil && ctx.Disabled[groupName] {
				continue
			}

			seg := registry.Get(groupName)
			if seg == nil {
				continue
			}

			out, err := seg.Render(ctx)
			if err != nil || out == nil {
				continue
			}

			segColor := themeColors[groupName]

			elements = append(elements, render.RenderedElement{
				Icon:       seg.DefaultIcon(styleMode),
				Title:      seg.DefaultTitle(),
				Primary:    out.Primary,
				Detail:     out.Detail,
				Color:      segColor.Fg,
				BgColor:    segColor.Bg,
				Priority:   seg.DefaultPriority(),
				IsVariable: out.IsVariable,
				MinWidth:   out.MinWidth,
			})
		}

		allElements = append(allElements, elements)
	}

	return allElements
}

package segment

import "strings"

// TokensGroup composes tokens_input, tokens_output, tokens_cached,
// and tokens_total segments with formatted separators.
type TokensGroup struct{}

func (g *TokensGroup) Name() string         { return "tokens" }
func (g *TokensGroup) DefaultTitle() string { return "" }
func (g *TokensGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("tokens", m)
}
func (g *TokensGroup) DefaultPriority() int { return 5 }

func (g *TokensGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	// Input and output are joined with "/": "12.3k↑/8.1k↓"
	var ioParts []string
	if !ctx.IsDisabled("tokens", "input") {
		seg := &TokensInputSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			ioParts = append(ioParts, out.Primary+"↑")
		}
	}
	if !ctx.IsDisabled("tokens", "output") {
		seg := &TokensOutputSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			ioParts = append(ioParts, out.Primary+"↓")
		}
	}

	// Cached and total are joined with "/": "2.1k♻/22.5k"
	var extraParts []string
	if !ctx.IsDisabled("tokens", "cached") {
		seg := &TokensCachedSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			extraParts = append(extraParts, out.Primary+"♻")
		}
	}
	if !ctx.IsDisabled("tokens", "total") {
		seg := &TokensTotalSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			extraParts = append(extraParts, out.Primary)
		}
	}

	var parts []string
	if len(ioParts) > 0 {
		parts = append(parts, strings.Join(ioParts, "/"))
	}
	if len(extraParts) > 0 {
		parts = append(parts, strings.Join(extraParts, "/"))
	}

	if len(parts) == 0 {
		return nil, nil
	}
	return &SegmentOutput{Primary: strings.Join(parts, " ")}, nil
}

package segment

import (
	"fmt"
)

// ContextPctSegment displays the context window used_percentage as-is.
type ContextPctSegment struct{}

func (s *ContextPctSegment) Name() string         { return "context_pct" }
func (s *ContextPctSegment) DefaultTitle() string { return "" }
func (s *ContextPctSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("context_pct", m)
}
func (s *ContextPctSegment) DefaultPriority() int { return 6 }

func (s *ContextPctSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.ContextWindow == nil {
		return nil, nil
	}
	return &SegmentOutput{
		Primary: fmt.Sprintf("%d%%", stdin.ContextWindow.UsedPercentage),
	}, nil
}

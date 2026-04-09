package segment

import (
	"github.com/ming-claude/ccsl/internal/render"
)

// ContextBarSegment displays a progress bar for context window usage.
type ContextBarSegment struct{}

func (s *ContextBarSegment) Name() string         { return "context_bar" }
func (s *ContextBarSegment) DefaultTitle() string { return "" }
func (s *ContextBarSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("context_bar", m)
}
func (s *ContextBarSegment) DefaultPriority() int { return 10 }

func (s *ContextBarSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.ContextWindow == nil {
		return nil, nil
	}
	bar := render.ProgressBar(stdin.ContextWindow.UsedPercentage, 10)
	return &SegmentOutput{Primary: bar}, nil
}

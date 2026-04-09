package segment

import (
	"fmt"
)

// LinesRemovedSegment displays the total lines removed in the session.
type LinesRemovedSegment struct{}

func (s *LinesRemovedSegment) Name() string         { return "lines_removed" }
func (s *LinesRemovedSegment) DefaultTitle() string { return "" }
func (s *LinesRemovedSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("lines_removed", m)
}
func (s *LinesRemovedSegment) DefaultPriority() int { return 2 }

func (s *LinesRemovedSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.Cost == nil {
		return nil, nil
	}
	return &SegmentOutput{Primary: fmt.Sprintf("%d", stdin.Cost.TotalLinesRemoved)}, nil
}

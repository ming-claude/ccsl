package segment

import (
	"fmt"
)

// LinesAddedSegment displays the total lines added in the session.
type LinesAddedSegment struct{}

func (s *LinesAddedSegment) Name() string         { return "lines_added" }
func (s *LinesAddedSegment) DefaultTitle() string { return "" }
func (s *LinesAddedSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("lines_added", m)
}
func (s *LinesAddedSegment) DefaultPriority() int { return 2 }

func (s *LinesAddedSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.Cost == nil {
		return nil, nil
	}
	return &SegmentOutput{Primary: fmt.Sprintf("%d", stdin.Cost.TotalLinesAdded)}, nil
}

package segment

import (
	"fmt"
)

// SessionCostSegment displays the total session cost in USD.
type SessionCostSegment struct{}

func (s *SessionCostSegment) Name() string         { return "session_cost" }
func (s *SessionCostSegment) DefaultTitle() string { return "" }
func (s *SessionCostSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("session_cost", m)
}
func (s *SessionCostSegment) DefaultPriority() int { return 8 }

func (s *SessionCostSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.Cost == nil {
		return nil, nil
	}
	cost := stdin.Cost.TotalCostUSD
	if cost == 0 {
		return &SegmentOutput{Primary: "$0"}, nil
	}
	return &SegmentOutput{Primary: fmt.Sprintf("$%.2f", cost)}, nil
}

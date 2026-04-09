package segment

import (
	"fmt"
)

// OutputSpeedSegment displays the session-average output token throughput.
type OutputSpeedSegment struct{}

func (s *OutputSpeedSegment) Name() string         { return "output_speed" }
func (s *OutputSpeedSegment) DefaultTitle() string { return "" }
func (s *OutputSpeedSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("output_speed", m)
}
func (s *OutputSpeedSegment) DefaultPriority() int { return 6 }

func (s *OutputSpeedSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.Cost == nil || stdin.ContextWindow == nil {
		return nil, nil
	}
	if stdin.Cost.TotalAPIDurationMs == 0 {
		return nil, nil
	}
	tokens := float64(stdin.ContextWindow.TotalOutputTokens)
	seconds := float64(stdin.Cost.TotalAPIDurationMs) / 1000.0
	speed := tokens / seconds
	return &SegmentOutput{Primary: formatSpeed(speed)}, nil
}

// formatSpeed formats a token/s rate as "X tok/s" or "X.Xk tok/s".
func formatSpeed(tokPerSec float64) string {
	if tokPerSec >= 1000 {
		return fmt.Sprintf("%.1fk tok/s", tokPerSec/1000)
	}
	return fmt.Sprintf("%.0f tok/s", tokPerSec)
}

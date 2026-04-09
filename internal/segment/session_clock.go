package segment

import (
	"fmt"
)

// SessionClockSegment displays session duration formatted as "Xh Xm", "Xm Xs", or "Xs".
type SessionClockSegment struct{}

func (s *SessionClockSegment) Name() string         { return "session_clock" }
func (s *SessionClockSegment) DefaultTitle() string { return "" }
func (s *SessionClockSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("session_clock", m)
}
func (s *SessionClockSegment) DefaultPriority() int { return 6 }

func (s *SessionClockSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.Cost == nil {
		return nil, nil
	}
	ms := stdin.Cost.TotalDurationMs
	if ms <= 0 {
		return nil, nil
	}
	return &SegmentOutput{Primary: formatDuration(ms)}, nil
}

func formatDuration(ms int64) string {
	totalSec := ms / 1000
	hours := totalSec / 3600
	minutes := (totalSec % 3600) / 60
	seconds := totalSec % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

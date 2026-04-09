package segment

import "time"

// ClockSegment displays the current time (HH:MM:SS) when the statusline is rendered.
type ClockSegment struct{}

func (s *ClockSegment) Name() string         { return "clock" }
func (s *ClockSegment) DefaultTitle() string { return "" }
func (s *ClockSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("clock", m)
}
func (s *ClockSegment) DefaultPriority() int { return 3 }

func (s *ClockSegment) Render(_ *RenderContext) (*SegmentOutput, error) {
	return &SegmentOutput{Primary: time.Now().Format("15:04:05")}, nil
}

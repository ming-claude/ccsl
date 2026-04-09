package segment

import (
	"github.com/ming-claude/ccsl/internal/input"
)

// WeeklyUsageSegment displays the 7-day rate limit usage as a percentage
// and progress bar. Data is sourced from stdin rate_limits.
type WeeklyUsageSegment struct{}

func (s *WeeklyUsageSegment) Name() string                   { return "weekly_usage" }
func (s *WeeklyUsageSegment) DefaultTitle() string           { return "7d" }
func (s *WeeklyUsageSegment) DefaultIcon(_ StyleMode) string { return "" }
func (s *WeeklyUsageSegment) DefaultPriority() int           { return 8 }

func (s *WeeklyUsageSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	return renderUsagePeriod(ctx,
		func(rl *input.RateLimits) *input.RateLimit { return rl.SevenDay },
		"\U000F1AE2", // nf-md-timer_sand_complete (7d)
	)
}

package segment

import (
	"fmt"
	"time"

	"github.com/ming-claude/ccsl/internal/input"
	"github.com/ming-claude/ccsl/internal/render"
)

// SessionUsageSegment displays the 5-hour rate limit usage as a percentage
// and progress bar. Data is sourced from stdin rate_limits.
type SessionUsageSegment struct{}

func (s *SessionUsageSegment) Name() string                   { return "session_usage" }
func (s *SessionUsageSegment) DefaultTitle() string           { return "5h" }
func (s *SessionUsageSegment) DefaultIcon(_ StyleMode) string { return "" }
func (s *SessionUsageSegment) DefaultPriority() int           { return 10 }

func (s *SessionUsageSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	return renderUsagePeriod(ctx,
		func(rl *input.RateLimits) *input.RateLimit { return rl.FiveHour },
		"\U000F13AB", // nf-md-timer_sand (5h)
	)
}

// renderUsagePeriod is the shared logic for SessionUsageSegment and
// WeeklyUsageSegment. It extracts rate limit data from stdin, showing a
// loading indicator when data is not yet available.
func renderUsagePeriod(
	ctx *RenderContext,
	stdinRL func(*input.RateLimits) *input.RateLimit,
	timerIcon string,
) (*SegmentOutput, error) {
	if ctx.Stdin == nil {
		return nil, nil
	}

	if ctx.Stdin.RateLimits == nil {
		return &SegmentOutput{Primary: "loading..."}, nil
	}

	rl := stdinRL(ctx.Stdin.RateLimits)
	if rl == nil {
		return &SegmentOutput{Primary: "loading..."}, nil
	}

	stale := isRateLimitOutdated(ctx, rl)

	pct := rl.UsedPercentage
	var primary string
	if stale {
		primary = fmt.Sprintf("%s \U000F02DA %.0f%%", render.ProgressBar(int(pct), 10), pct)
	} else {
		primary = fmt.Sprintf("%s %.0f%%", render.ProgressBar(int(pct), 10), pct)
	}
	if rl.ResetsAt > 0 {
		if remaining := time.Until(time.Unix(rl.ResetsAt, 0)); remaining > 0 {
			primary += " " + timerIcon + " " + formatCountdown(remaining)
		}
	}

	return &SegmentOutput{Primary: primary}, nil
}

// rateLimitStaleThreshold is the maximum age of transcript data before
// rate-limit percentages are considered outdated.
const rateLimitStaleThreshold = 10 * time.Minute

// isRateLimitOutdated returns true when the rate-limit data should no
// longer be trusted. Two conditions trigger this:
//  1. The rate-limit window has already reset (resets_at in the past).
//  2. The last API response is older than rateLimitStaleThreshold,
//     meaning the underlying data snapshot is too old to be meaningful.
func isRateLimitOutdated(ctx *RenderContext, rl *input.RateLimit) bool {
	// Window already expired — data is from a previous period.
	if rl.ResetsAt > 0 && time.Now().After(time.Unix(rl.ResetsAt, 0)) {
		return true
	}

	// No transcript or no events — can't determine freshness, trust stdin.
	if ctx.Transcript == nil || ctx.Transcript.LastResponseTime.IsZero() {
		return false
	}

	return time.Since(ctx.Transcript.LastResponseTime) > rateLimitStaleThreshold
}

package segment

// MemoryUsageSegment displays the system memory usage percentage.
// The value is pre-fetched from vm_stat during the concurrent fetch
// phase and stored in RenderContext.MemoryPct.
type MemoryUsageSegment struct{}

func (s *MemoryUsageSegment) Name() string         { return "memory_usage" }
func (s *MemoryUsageSegment) DefaultTitle() string { return "" }
func (s *MemoryUsageSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("memory_usage", m)
}
func (s *MemoryUsageSegment) DefaultPriority() int { return 2 }

func (s *MemoryUsageSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	if ctx.MemoryPct == "" {
		return nil, nil
	}
	return &SegmentOutput{Primary: ctx.MemoryPct}, nil
}

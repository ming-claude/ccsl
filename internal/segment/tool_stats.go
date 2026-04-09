package segment

import (
	"fmt"
	"sort"
	"strings"
)

// ToolStatsSegment displays per-tool invocation counts, sorted by frequency.
type ToolStatsSegment struct{}

func (s *ToolStatsSegment) Name() string         { return "tool_stats" }
func (s *ToolStatsSegment) DefaultTitle() string { return "" }
func (s *ToolStatsSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("tool_stats", m)
}
func (s *ToolStatsSegment) DefaultPriority() int { return 4 }

func (s *ToolStatsSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	td := ctx.Transcript
	if td == nil || len(td.ToolStats) == 0 {
		return nil, nil
	}

	type entry struct {
		Name  string
		Count int
	}

	entries := make([]entry, 0, len(td.ToolStats))
	for name, count := range td.ToolStats {
		entries = append(entries, entry{Name: name, Count: count})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Count != entries[j].Count {
			return entries[i].Count > entries[j].Count
		}
		return entries[i].Name < entries[j].Name
	})

	parts := make([]string, len(entries))
	for i, e := range entries {
		parts[i] = fmt.Sprintf("%s \u00d7%d", e.Name, e.Count)
	}

	return &SegmentOutput{Primary: strings.Join(parts, "  ")}, nil
}

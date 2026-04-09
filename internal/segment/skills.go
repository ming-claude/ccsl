package segment

import (
	"fmt"
)

// SkillsSegment displays the most recent skill name and total count.
type SkillsSegment struct{}

func (s *SkillsSegment) Name() string         { return "skills" }
func (s *SkillsSegment) DefaultTitle() string { return "" }
func (s *SkillsSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("skills", m)
}
func (s *SkillsSegment) DefaultPriority() int { return 2 }

func (s *SkillsSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	td := ctx.Transcript
	if td == nil || len(td.Skills) == 0 {
		return nil, nil
	}

	latest := td.Skills[len(td.Skills)-1]
	if len(td.Skills) > 1 {
		return &SegmentOutput{Primary: fmt.Sprintf("%s (%d)", latest, len(td.Skills))}, nil
	}

	return &SegmentOutput{Primary: latest}, nil
}

package segment

import (
	"fmt"
	"strings"
)

// GitChangesSegment displays a summary of uncommitted file changes.
type GitChangesSegment struct{}

func (s *GitChangesSegment) Name() string         { return "git_changes" }
func (s *GitChangesSegment) DefaultTitle() string { return "" }
func (s *GitChangesSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("git_changes", m)
}
func (s *GitChangesSegment) DefaultPriority() int { return 6 }

func (s *GitChangesSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	gitData := ctx.Git
	if gitData == nil || gitData.IsClean {
		return nil, nil
	}

	var parts []string
	if gitData.Added > 0 {
		parts = append(parts, fmt.Sprintf("+%d", gitData.Added))
	}
	if gitData.Modified > 0 {
		parts = append(parts, fmt.Sprintf("~%d", gitData.Modified))
	}
	if gitData.Deleted > 0 {
		parts = append(parts, fmt.Sprintf("-%d", gitData.Deleted))
	}
	if gitData.Untracked > 0 {
		parts = append(parts, fmt.Sprintf("?%d", gitData.Untracked))
	}

	if len(parts) == 0 {
		return nil, nil
	}

	return &SegmentOutput{Primary: strings.Join(parts, " ")}, nil
}

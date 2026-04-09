package segment

import (
	"fmt"
)

// GitDeletionsSegment displays the number of deleted lines in the working tree.
type GitDeletionsSegment struct{}

func (s *GitDeletionsSegment) Name() string         { return "git_deletions" }
func (s *GitDeletionsSegment) DefaultTitle() string { return "" }
func (s *GitDeletionsSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("git_deletions", m)
}
func (s *GitDeletionsSegment) DefaultPriority() int { return 6 }

func (s *GitDeletionsSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	gitData := ctx.Git
	if gitData == nil || gitData.Deletions == 0 {
		return nil, nil
	}
	return &SegmentOutput{Primary: fmt.Sprintf("-%d", gitData.Deletions)}, nil
}

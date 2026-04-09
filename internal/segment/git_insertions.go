package segment

import (
	"fmt"
)

// GitInsertionsSegment displays the number of inserted lines in the working tree.
type GitInsertionsSegment struct{}

func (s *GitInsertionsSegment) Name() string         { return "git_insertions" }
func (s *GitInsertionsSegment) DefaultTitle() string { return "" }
func (s *GitInsertionsSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("git_insertions", m)
}
func (s *GitInsertionsSegment) DefaultPriority() int { return 6 }

func (s *GitInsertionsSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	gitData := ctx.Git
	if gitData == nil || gitData.Insertions == 0 {
		return nil, nil
	}
	return &SegmentOutput{Primary: fmt.Sprintf("+%d", gitData.Insertions)}, nil
}

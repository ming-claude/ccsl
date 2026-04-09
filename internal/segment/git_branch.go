package segment

import (
	"fmt"
	"strings"
)

// GitBranchSegment displays the current git branch name and ahead/behind counts.
type GitBranchSegment struct{}

func (s *GitBranchSegment) Name() string         { return "git_branch" }
func (s *GitBranchSegment) DefaultTitle() string { return "" }
func (s *GitBranchSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("git_branch", m)
}
func (s *GitBranchSegment) DefaultPriority() int { return 8 }

func (s *GitBranchSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	gitData := ctx.Git
	if gitData == nil || gitData.Branch == "" {
		return nil, nil
	}

	out := &SegmentOutput{Primary: gitData.Branch}

	if gitData.HasRemote {
		var parts []string
		if gitData.Ahead > 0 {
			parts = append(parts, fmt.Sprintf("↑%d", gitData.Ahead))
		}
		if gitData.Behind > 0 {
			parts = append(parts, fmt.Sprintf("↓%d", gitData.Behind))
		}
		if len(parts) > 0 {
			out.Detail = strings.Join(parts, " ")
		}
	}

	return out, nil
}

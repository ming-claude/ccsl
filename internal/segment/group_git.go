package segment

import "strings"

// GitGroup composes git_branch, git_changes, git_insertions, git_deletions,
// and git_worktree segments.
type GitGroup struct{}

func (g *GitGroup) Name() string         { return "git" }
func (g *GitGroup) DefaultTitle() string { return "" }
func (g *GitGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("git", m)
}
func (g *GitGroup) DefaultPriority() int { return 6 }

func (g *GitGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	var parts []string

	// Branch is always first.
	if !ctx.IsDisabled("git", "branch") {
		seg := &GitBranchSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			text := out.Primary
			if out.Detail != "" {
				text += " " + out.Detail
			}
			parts = append(parts, text)
		}
	}

	// Worktree after branch.
	if !ctx.IsDisabled("git", "worktree") {
		seg := &GitWorktreeSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	// File-level changes.
	if !ctx.IsDisabled("git", "changes") {
		seg := &GitChangesSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	// Insertions/deletions formatted as "+N/-N".
	var diffParts []string
	if !ctx.IsDisabled("git", "insertions") {
		seg := &GitInsertionsSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			diffParts = append(diffParts, out.Primary)
		}
	}
	if !ctx.IsDisabled("git", "deletions") {
		seg := &GitDeletionsSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			diffParts = append(diffParts, out.Primary)
		}
	}
	if len(diffParts) > 0 {
		parts = append(parts, strings.Join(diffParts, "/"))
	}

	if len(parts) == 0 {
		return nil, nil
	}
	return &SegmentOutput{
		Primary:    strings.Join(parts, " "),
		IsVariable: true,
		MinWidth:   15,
	}, nil
}

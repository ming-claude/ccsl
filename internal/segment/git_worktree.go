package segment

// GitWorktreeSegment displays the worktree name and branch.
type GitWorktreeSegment struct{}

func (s *GitWorktreeSegment) Name() string         { return "git_worktree" }
func (s *GitWorktreeSegment) DefaultTitle() string { return "" }
func (s *GitWorktreeSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("git_worktree", m)
}
func (s *GitWorktreeSegment) DefaultPriority() int { return 4 }

func (s *GitWorktreeSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.Worktree == nil {
		return nil, nil
	}
	wt := stdin.Worktree
	if wt.Name == "" {
		return nil, nil
	}
	text := wt.Name
	if wt.Branch != "" {
		text += " (" + wt.Branch + ")"
	}
	return &SegmentOutput{Primary: text}, nil
}

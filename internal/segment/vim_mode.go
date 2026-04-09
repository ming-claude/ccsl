package segment

// VimModeSegment displays the current Vim mode.
type VimModeSegment struct{}

func (s *VimModeSegment) Name() string         { return "vim_mode" }
func (s *VimModeSegment) DefaultTitle() string { return "" }
func (s *VimModeSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("vim_mode", m)
}
func (s *VimModeSegment) DefaultPriority() int { return 4 }

func (s *VimModeSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.Vim == nil {
		return nil, nil
	}
	if stdin.Vim.Mode == "" {
		return nil, nil
	}
	return &SegmentOutput{Primary: stdin.Vim.Mode}, nil
}

package segment

// OutputStyleSegment displays the current output style name.
type OutputStyleSegment struct{}

func (s *OutputStyleSegment) Name() string         { return "output_style" }
func (s *OutputStyleSegment) DefaultTitle() string { return "" }
func (s *OutputStyleSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("output_style", m)
}
func (s *OutputStyleSegment) DefaultPriority() int { return 4 }

func (s *OutputStyleSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.OutputStyle == nil {
		return nil, nil
	}
	if stdin.OutputStyle.Name == "" {
		return nil, nil
	}
	return &SegmentOutput{Primary: stdin.OutputStyle.Name}, nil
}

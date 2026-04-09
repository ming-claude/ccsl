package segment

// ModelSegment displays the model display name from stdin.
// e.g. "Opus 4.6 (1M context)"
type ModelSegment struct{}

func (s *ModelSegment) Name() string         { return "model" }
func (s *ModelSegment) DefaultTitle() string { return "" }
func (s *ModelSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("model", m)
}
func (s *ModelSegment) DefaultPriority() int { return 10 }

func (s *ModelSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.Model == nil {
		return nil, nil
	}
	name := stdin.Model.DisplayName
	if name == "" {
		name = stdin.Model.ID
	}
	return &SegmentOutput{Primary: name}, nil
}

package segment

// SessionNameSegment displays the session name.
type SessionNameSegment struct{}

func (s *SessionNameSegment) Name() string         { return "session_name" }
func (s *SessionNameSegment) DefaultTitle() string { return "" }
func (s *SessionNameSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("session_name", m)
}
func (s *SessionNameSegment) DefaultPriority() int { return 4 }

func (s *SessionNameSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.SessionName == "" {
		return nil, nil
	}
	return &SegmentOutput{Primary: stdin.SessionName}, nil
}

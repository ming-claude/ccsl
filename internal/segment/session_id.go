package segment

// SessionIDSegment displays the first 8 characters of the session ID.
type SessionIDSegment struct{}

func (s *SessionIDSegment) Name() string         { return "session_id" }
func (s *SessionIDSegment) DefaultTitle() string { return "" }
func (s *SessionIDSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("session_id", m)
}
func (s *SessionIDSegment) DefaultPriority() int { return 2 }

func (s *SessionIDSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.SessionID == "" {
		return nil, nil
	}
	id := stdin.SessionID
	if len(id) > 8 {
		id = id[:8]
	}
	return &SegmentOutput{Primary: id}, nil
}

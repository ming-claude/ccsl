package segment

import "strings"

// SessionGroup composes session_id, session_name, and session_clock segments.
type SessionGroup struct{}

func (g *SessionGroup) Name() string         { return "session" }
func (g *SessionGroup) DefaultTitle() string { return "" }
func (g *SessionGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("session", m)
}
func (g *SessionGroup) DefaultPriority() int { return 5 }

func (g *SessionGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	var parts []string

	if !ctx.IsDisabled("session", "id") {
		seg := &SessionIDSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if !ctx.IsDisabled("session", "name") {
		seg := &SessionNameSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if !ctx.IsDisabled("session", "clock") {
		seg := &SessionClockSegment{}
		if out, _ := seg.Render(ctx); out != nil {
			parts = append(parts, out.Primary)
		}
	}

	if len(parts) == 0 {
		return nil, nil
	}
	return &SegmentOutput{Primary: strings.Join(parts, " ")}, nil
}

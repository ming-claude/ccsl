package segment

import (
	"strings"
)

// CWDSegment displays the last N path components of the current working directory.
type CWDSegment struct{}

func (s *CWDSegment) Name() string         { return "cwd" }
func (s *CWDSegment) DefaultTitle() string { return "" }
func (s *CWDSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("cwd", m)
}
func (s *CWDSegment) DefaultPriority() int { return 4 }

func (s *CWDSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.CWD == "" {
		return nil, nil
	}
	return &SegmentOutput{
		Primary:    lastNComponents(stdin.CWD, 2),
		IsVariable: true,
		MinWidth:   12,
	}, nil
}

// lastNComponents returns the last n path components of a path.
func lastNComponents(path string, n int) string {
	// Clean trailing slash.
	path = strings.TrimRight(path, "/")
	if path == "" {
		return "/"
	}
	parts := strings.Split(path, "/")
	if n >= len(parts) {
		return path
	}
	return strings.Join(parts[len(parts)-n:], "/")
}

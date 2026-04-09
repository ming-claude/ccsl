package segment

import (
	"fmt"

	"github.com/ming-claude/ccsl/internal/render"
)

// TerminalWidthSegment displays the current terminal width in columns.
type TerminalWidthSegment struct{}

func (s *TerminalWidthSegment) Name() string         { return "terminal_width" }
func (s *TerminalWidthSegment) DefaultTitle() string { return "" }
func (s *TerminalWidthSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("terminal_width", m)
}
func (s *TerminalWidthSegment) DefaultPriority() int { return 2 }

func (s *TerminalWidthSegment) Render(_ *RenderContext) (*SegmentOutput, error) {
	width := render.TerminalWidth(80)
	return &SegmentOutput{Primary: fmt.Sprintf("%d", width)}, nil
}

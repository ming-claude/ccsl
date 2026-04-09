package segment

import (
	gitpkg "github.com/ming-claude/ccsl/internal/git"
	"github.com/ming-claude/ccsl/internal/input"
	"github.com/ming-claude/ccsl/internal/npm"
	"github.com/ming-claude/ccsl/internal/transcript"
)

// StyleMode represents the icon/glyph style for segment rendering.
type StyleMode string

const (
	StylePlain     StyleMode = "plain"
	StyleNerdFont  StyleMode = "nerd-font"
	StylePowerline StyleMode = "powerline"
)

// SegmentOutput holds the rendered output of a segment.
type SegmentOutput struct {
	Primary    string
	Detail     string
	IsVariable bool // can be truncated by the width budget algorithm
	MinWidth   int  // minimum display width before hiding (variable-width only)
}

// RenderContext holds all data sources available to segments during rendering.
type RenderContext struct {
	Stdin      *input.StdinData
	Transcript *transcript.TranscriptData
	Git        *gitpkg.GitData
	NpmVersion *npm.VersionInfo

	MemoryPct string // from vm_stat

	Style    StyleMode
	Disabled map[string]bool
}

// IsDisabled checks whether a child segment within a group is disabled.
// The key format is "group.child" (e.g., "git.branch").
func (ctx *RenderContext) IsDisabled(group, child string) bool {
	if ctx.Disabled == nil {
		return false
	}
	return ctx.Disabled[group+"."+child]
}

// Segment is the interface that every statusline segment must implement.
type Segment interface {
	Name() string
	Render(ctx *RenderContext) (*SegmentOutput, error)
	DefaultTitle() string
	DefaultIcon(style StyleMode) string
	DefaultPriority() int
}

// IconResolver is a function that returns the default icon for a segment
// in the given style mode. It is set by the style package to break the
// import cycle between segment and style.
var IconResolver func(segmentName string, mode StyleMode) string

// resolveIcon calls IconResolver if set, otherwise returns "".
func resolveIcon(segmentName string, mode StyleMode) string {
	if IconResolver != nil {
		return IconResolver(segmentName, mode)
	}
	return ""
}

// Registry stores named segments and preserves registration order.
type Registry struct {
	segments map[string]Segment
	order    []string
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		segments: make(map[string]Segment),
	}
}

// Register adds a segment to the registry. If a segment with the same name
// already exists, it is replaced (but the original insertion order is kept).
func (r *Registry) Register(s Segment) {
	name := s.Name()
	if _, exists := r.segments[name]; !exists {
		r.order = append(r.order, name)
	}
	r.segments[name] = s
}

// Get returns the segment with the given name, or nil if not found.
func (r *Registry) Get(name string) Segment {
	return r.segments[name]
}

// All returns all registered segments in registration order.
func (r *Registry) All() []Segment {
	out := make([]Segment, 0, len(r.order))
	for _, name := range r.order {
		out = append(out, r.segments[name])
	}
	return out
}

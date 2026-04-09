package config

import (
	gojson "github.com/goccy/go-json"
)

// UserConfig is the serialized user configuration stored on disk.
// It contains only user-facing knobs: preset selection, theme selection,
// style, and disabled segments.
type UserConfig struct {
	Preset           string   `json:"preset"`
	Theme            string   `json:"theme"`
	Style            string   `json:"style"`
	DisabledSegments []string `json:"disabled_segments"`
	Debug            bool     `json:"debug,omitempty"`
}

// Config is the resolved runtime configuration.
// It is built by Load() from a UserConfig + preset + theme and is read-only at runtime.
type Config struct {
	Preset       string
	Theme        string
	Style        string
	Lines        [][]string              // group name arrays per line, from preset
	Disabled     map[string]bool         // union of preset defaults + user disabled_segments
	ThemeColors  map[string]SegmentColor // from theme, keyed by segment/group name
	Separator    string                  // from theme
	SeparatorSet bool                    // true when theme provides a separator
	Debug        bool                    // enable debug logging
}

// SegmentColor represents a color specification for a segment or group.
// It supports two JSON forms:
//   - string: foreground color only, e.g. "#61afef" or "red"
//   - object: foreground + background + optional semantic roles,
//     e.g. {"fg": "#98d885", "bg": "#2d4530", "positive": "#98d885", "negative": "#e87070"}
type SegmentColor struct {
	Fg string `json:"fg,omitempty"`
	Bg string `json:"bg,omitempty"`
	// TODO: wire Positive/Negative through pipeline for semantic coloring (e.g., git insertions/deletions)
	Positive string `json:"positive,omitempty"` // semantic: green/success (e.g., git insertions)
	Negative string `json:"negative,omitempty"` // semantic: red/error (e.g., git deletions)
}

// IsZero reports whether the color is empty (no fg or bg set).
func (c SegmentColor) IsZero() bool {
	return c.Fg == "" && c.Bg == ""
}

func (c SegmentColor) MarshalJSON() ([]byte, error) {
	if c.Bg == "" && c.Positive == "" && c.Negative == "" {
		return gojson.Marshal(c.Fg)
	}
	type alias SegmentColor
	return gojson.Marshal(alias(c))
}

func (c *SegmentColor) UnmarshalJSON(data []byte) error {
	// Try string first (fg-only shorthand).
	var s string
	if err := gojson.Unmarshal(data, &s); err == nil {
		c.Fg = s
		c.Bg = ""
		c.Positive = ""
		c.Negative = ""
		return nil
	}
	// Try object {fg, bg, positive, negative}.
	type alias SegmentColor
	var a alias
	if err := gojson.Unmarshal(data, &a); err != nil {
		return err
	}
	*c = SegmentColor(a)
	return nil
}

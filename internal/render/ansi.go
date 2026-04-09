package render

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

const reset = "\x1b[0m"

// namedColors maps color names to their ANSI escape codes.
var namedColors = map[string]string{
	"red":     "\x1b[31m",
	"green":   "\x1b[32m",
	"yellow":  "\x1b[33m",
	"blue":    "\x1b[34m",
	"magenta": "\x1b[35m",
	"cyan":    "\x1b[36m",
	"white":   "\x1b[37m",
	"dim":     "\x1b[2m",
}

// colorCode returns the ANSI escape code for a color string.
// Supports named colors, hex colors (#rrggbb), and 256-color codes.
// isBg selects background (48) vs foreground (38) mode.
// Returns "" if the color is empty or invalid.
func colorCode(color string, isBg bool) string {
	if color == "" {
		return ""
	}

	base := 38
	if isBg {
		base = 48
	}

	switch {
	case namedColors[color] != "":
		if isBg {
			// Named colors: fg is 3x, bg is 4x.
			code := namedColors[color]
			return strings.Replace(code, "\x1b[3", "\x1b[4", 1)
		}
		return namedColors[color]
	case strings.HasPrefix(color, "#") && len(color) == 7:
		r, errR := strconv.ParseUint(color[1:3], 16, 8)
		g, errG := strconv.ParseUint(color[3:5], 16, 8)
		b, errB := strconv.ParseUint(color[5:7], 16, 8)
		if errR != nil || errG != nil || errB != nil {
			return ""
		}
		return fmt.Sprintf("\x1b[%d;2;%d;%d;%dm", base, r, g, b)
	default:
		n, err := strconv.Atoi(color)
		if err != nil || n < 0 || n > 255 {
			return ""
		}
		return fmt.Sprintf("\x1b[%d;5;%dm", base, n)
	}
}

// Colorize wraps text with ANSI foreground color codes. Supports named colors,
// hex colors (e.g. "#ff0000"), and 256-color codes (e.g. "123"). Empty text or
// color returns text unchanged.
func Colorize(text, color string) string {
	if text == "" || color == "" {
		return text
	}
	code := colorCode(color, false)
	if code == "" {
		return text
	}
	return code + text + reset
}

// ColorizeBlock wraps text with foreground and background colors, adding
// padding spaces on both sides for a block/badge appearance. When bg is
// empty, falls back to fg-only coloring via Colorize.
func ColorizeBlock(text, fg, bg string) string {
	if text == "" {
		return text
	}
	if bg == "" {
		return Colorize(text, fg)
	}
	bgCode := colorCode(bg, true)
	if bgCode == "" {
		return Colorize(text, fg)
	}
	fgCode := colorCode(fg, false)
	// Pad with spaces inside the bg block for badge look.
	return bgCode + fgCode + " " + text + " " + reset
}

// DisplayWidth returns the visual display width of s, ignoring ANSI escape
// sequences. CJK characters count as width 2.
func DisplayWidth(s string) int {
	width := 0
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			// Skip ANSI escape sequence: consume until a letter [a-zA-Z].
			i++
			for i < len(s) {
				b := s[i]
				i++
				if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') {
					break
				}
			}
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		if r == utf8.RuneError {
			continue
		}
		width += runeWidth(r)
	}
	return width
}

// runeWidth returns the display width of a single rune.
func runeWidth(r rune) int {
	return runewidth.RuneWidth(r)
}

// fractionalBlocks maps 1/8th increments to Unicode left-block characters.
// Index 0 = empty (no partial block), index 8 = full block.
var fractionalBlocks = [9]string{
	"", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "█",
}

// ProgressBar renders a borderless progress bar with 1/8th character
// precision via Unicode fractional block elements.
// percent is clamped to [0, 100]. width is the total display width.
// The empty portion uses spaces so the segment's ANSI background color
// serves as the track — fractional blocks blend seamlessly since their
// unfilled side also shows the background color.
func ProgressBar(percent, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	if width < 1 {
		return ""
	}

	// Total units in 1/8th character granularity.
	totalEighths := width * 8
	filledEighths := totalEighths * percent / 100

	fullBlocks := filledEighths / 8
	remainder := filledEighths % 8
	emptyBlocks := width - fullBlocks
	if remainder > 0 {
		emptyBlocks--
	}

	return strings.Repeat("█", fullBlocks) + fractionalBlocks[remainder] + strings.Repeat(" ", emptyBlocks)
}

// Truncate truncates s to maxWidth display columns, appending "…" if the
// string was truncated. Handles CJK characters properly by never splitting a
// double-width character. Reserves 1 column for the ellipsis when truncation
// is needed.
func Truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	totalWidth := DisplayWidth(s)
	if totalWidth <= maxWidth {
		return s
	}

	// Reserve 1 column for "…".
	limit := maxWidth - 1
	var b strings.Builder
	current := 0
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			// Preserve ANSI escapes — they have zero display width.
			start := i
			i++
			for i < len(s) {
				ch := s[i]
				i++
				if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
					break
				}
			}
			b.WriteString(s[start:i])
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		w := runeWidth(r)
		if current+w > limit {
			break
		}
		b.WriteRune(r)
		current += w
		i += size
	}

	b.WriteString("…")
	return b.String()
}

// StripANSI removes all ANSI escape sequences from s.
func StripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			i++
			for i < len(s) {
				ch := s[i]
				i++
				if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
					break
				}
			}
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

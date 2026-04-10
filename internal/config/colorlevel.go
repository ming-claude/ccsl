package config

import (
	"os"
	"strconv"
	"strings"
)

// Color level constants matching supports-color npm package semantics.
const (
	ColorLevelNone      = 0 // No color (NO_COLOR or TERM=dumb)
	ColorLevelBasic     = 1 // 16-color ANSI
	ColorLevel256       = 2 // 256-color palette
	ColorLevelTruecolor = 3 // 24-bit RGB
)

// DetectColorLevel returns the terminal's color support level (0–3)
// by inspecting environment variables, mirroring the supports-color
// npm package used by chalk in Claude Code's Ink TUI.
func DetectColorLevel() int {
	return detectColorLevelFromEnv(os.Getenv)
}

// detectColorLevelFromEnv is the testable core. The getenv parameter
// allows tests to inject environment without mutating os.Environ.
func detectColorLevelFromEnv(getenv func(string) string) int {
	term := getenv("TERM")

	// FORCE_COLOR overrides everything (supports-color compat).
	if fc := getenv("FORCE_COLOR"); fc != "" {
		n, err := strconv.Atoi(fc)
		if err != nil || n < 0 {
			return ColorLevelBasic
		}
		if n > 3 {
			return ColorLevelTruecolor
		}
		return n
	}

	// NO_COLOR or TERM=dumb → no color.
	if getenv("NO_COLOR") != "" {
		return ColorLevelNone
	}
	if term == "dumb" {
		return ColorLevelNone
	}

	// COLORTERM=truecolor or 24bit → level 3.
	ct := getenv("COLORTERM")
	if ct == "truecolor" || ct == "24bit" {
		return ColorLevelTruecolor
	}

	// Known truecolor terminals identified by TERM value.
	termLower := strings.ToLower(term)
	for _, kw := range []string{"ghostty", "kitty", "wezterm"} {
		if strings.Contains(termLower, kw) {
			return ColorLevelTruecolor
		}
	}

	// 256-color detection.
	if strings.Contains(termLower, "256color") {
		return ColorLevel256
	}

	// Basic terminal patterns.
	for _, prefix := range []string{
		"xterm", "vt100", "vt220", "screen", "tmux",
		"linux", "rxvt", "cygwin", "ansi",
	} {
		if strings.HasPrefix(termLower, prefix) {
			return ColorLevelBasic
		}
	}

	return ColorLevelBasic
}

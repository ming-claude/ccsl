package render

import "testing"

func TestColorize(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		color string
		want  string
	}{
		{"named red", "hello", "red", "\x1b[31mhello\x1b[0m"},
		{"named green", "ok", "green", "\x1b[32mok\x1b[0m"},
		{"named dim", "faint", "dim", "\x1b[2mfaint\x1b[0m"},
		{"hex color", "hi", "#ff0000", "\x1b[38;2;255;0;0mhi\x1b[0m"},
		{"hex mixed case", "hi", "#00ff00", "\x1b[38;2;0;255;0mhi\x1b[0m"},
		{"256 color", "x", "123", "\x1b[38;5;123mx\x1b[0m"},
		{"256 color zero", "x", "0", "\x1b[38;5;0mx\x1b[0m"},
		{"empty text", "", "red", ""},
		{"empty color", "hello", "", "hello"},
		{"both empty", "", "", ""},
		{"invalid color name", "hi", "nope", "hi"},
		{"invalid hex", "hi", "#zzzzzz", "hi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Colorize(tt.text, tt.color)
			if got != tt.want {
				t.Errorf("Colorize(%q, %q) = %q, want %q", tt.text, tt.color, got, tt.want)
			}
		})
	}
}

func TestRuneWidth_ZeroWidth(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		{"zero-width joiner", '\u200D', 0},
		{"zero-width non-joiner", '\u200C', 0},
		{"combining acute accent", '\u0301', 0},
		{"regular ASCII", 'A', 1},
		{"CJK character", '你', 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runeWidth(tt.r)
			if got != tt.want {
				t.Errorf("runeWidth(%U) = %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"ascii", "hello", 5},
		{"CJK", "你好", 4},
		{"mixed", "hi你好", 6},
		{"with ANSI", "\x1b[31mhello\x1b[0m", 5},
		{"empty", "", 0},
		{"CJK with ANSI", "\x1b[32m你好\x1b[0m", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DisplayWidth(tt.s)
			if got != tt.want {
				t.Errorf("DisplayWidth(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestProgressBar(t *testing.T) {
	tests := []struct {
		name    string
		percent int
		width   int
		want    string
	}{
		// Borderless: width = total bar columns (no caps).
		{"50 percent w10", 50, 10, "█████     "},
		{"0 percent w10", 0, 10, "          "},
		{"100 percent w10", 100, 10, "██████████"},
		{"negative clamped", -10, 10, "          "},
		{"over 100 clamped", 150, 10, "██████████"},
		{"too small width", 50, 0, ""},
		// Fractional block precision.
		{"16 percent w10", 16, 10, "█▌        "},
		{"5 percent w10", 5, 10, "▌         "},
		{"99 percent w10", 99, 10, "█████████▉"},
		{"33 percent w10", 33, 10, "███▎      "},
		{"1 percent w10", 1, 10, "          "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProgressBar(tt.percent, tt.width)
			if got != tt.want {
				t.Errorf("ProgressBar(%d, %d) = %q, want %q", tt.percent, tt.width, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxWidth int
		want     string
	}{
		{"truncated", "hello world", 8, "hello w…"},
		{"fits exactly", "hello", 5, "hello"},
		{"shorter", "hi", 10, "hi"},
		{"CJK truncation", "你好世界", 5, "你好…"},
		{"CJK no split", "你好世界", 4, "你…"},
		{"empty string", "", 5, ""},
		{"zero width", "hello", 0, ""},
		{"width 1", "hello", 1, "…"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.s, tt.maxWidth)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.s, tt.maxWidth, got, tt.want)
			}
		})
	}
}

func TestColorizeBlock(t *testing.T) {
	tests := []struct {
		name string
		text string
		fg   string
		bg   string
		want string
	}{
		{
			"fg+bg hex",
			"hello",
			"#282c34",
			"#c678dd",
			"\x1b[48;2;198;120;221m\x1b[38;2;40;44;52m hello \x1b[0m",
		},
		{
			"no bg falls back to fg-only",
			"hello",
			"red",
			"",
			"\x1b[31mhello\x1b[0m",
		},
		{
			"empty text",
			"",
			"red",
			"blue",
			"",
		},
		{
			"bg only, no fg",
			"hello",
			"",
			"#c678dd",
			"\x1b[48;2;198;120;221m hello \x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ColorizeBlock(tt.text, tt.fg, tt.bg)
			if got != tt.want {
				t.Errorf("ColorizeBlock(%q, %q, %q) = %q, want %q", tt.text, tt.fg, tt.bg, got, tt.want)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"colored text", "\x1b[31mhello\x1b[0m", "hello"},
		{"bold text", "\x1b[1mworld\x1b[0m", "world"},
		{"no ANSI", "plain", "plain"},
		{"empty", "", ""},
		{"multiple escapes", "\x1b[1m\x1b[31mhi\x1b[0m", "hi"},
		{"256 color", "\x1b[38;5;123mtest\x1b[0m", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripANSI(tt.s)
			if got != tt.want {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

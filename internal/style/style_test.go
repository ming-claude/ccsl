package style

import (
	"testing"

	"github.com/ming-claude/ccsl/internal/segment"
)

func TestDefaultIcon(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		mode      segment.StyleMode
		want      string
	}{
		{"plain model", "model", segment.StylePlain, ">"},
		{"nerd-font model", "model", segment.StyleNerdFont, "\U000F167A"}, // nf-md-head_cog
		{"plain git", "git", segment.StylePlain, "*"},
		{"nerd-font git", "git", segment.StyleNerdFont, "\U000F02A2"}, // nf-md-git
		{"plain context", "context", segment.StylePlain, "Ctx:"},
		{"nerd-font context", "context", segment.StyleNerdFont, "\U000F035B"}, // nf-md-chip
		{"plain tokens", "tokens", segment.StylePlain, "Tok:"},
		{"plain cost", "cost", segment.StylePlain, "$"},
		{"nerd-font cost", "cost", segment.StyleNerdFont, "\uef8d"}, // nf-fa-sack_dollar
		{"plain usage_5hour", "usage_5hour", segment.StylePlain, "%:"},
		{"nerd-font usage_5hour", "usage_5hour", segment.StyleNerdFont, "\uf200"}, // nf-fa-chart_pie
		{"plain session", "session", segment.StylePlain, "S:"},
		{"nerd-font session", "session", segment.StyleNerdFont, "\U000F0220"}, // nf-md-clock_outline
		{"plain speed", "speed", segment.StylePlain, "Spd:"},
		{"nerd-font speed", "speed", segment.StyleNerdFont, "\U000F04C5"}, // nf-md-rocket
		{"plain diff", "diff", segment.StylePlain, "+-:"},
		{"nerd-font diff", "diff", segment.StyleNerdFont, "\U000F0992"}, // nf-md-compare
		{"plain activity", "activity", segment.StylePlain, "~:"},
		{"nerd-font activity", "activity", segment.StyleNerdFont, "\U000F0211"}, // nf-md-fire
		{"plain cwd", "cwd", segment.StylePlain, "~/"},
		{"nerd-font cwd", "cwd", segment.StyleNerdFont, "\U000F0770"}, // nf-md-folder_open
		{"plain version", "version", segment.StylePlain, "v:"},
		{"nerd-font version", "version", segment.StyleNerdFont, "\uf409"}, // nf-oct-download
		{"plain env", "env", segment.StylePlain, "Env:"},
		{"nerd-font env", "env", segment.StyleNerdFont, "\U000F0635"}, // nf-md-shield
		{"unknown group", "nonexistent", segment.StylePlain, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultIcon(tt.groupName, tt.mode)
			if got != tt.want {
				t.Errorf("DefaultIcon(%q, %q) = %q, want %q", tt.groupName, tt.mode, got, tt.want)
			}
		})
	}
}

func TestDefaultIcon_AllGroupsCovered(t *testing.T) {
	// Verify all 14 groups have entries for all 3 modes.
	expectedGroups := []string{
		"model", "git", "context", "tokens", "cost", "usage_5hour", "usage_weekly", "version",
		"session", "speed", "diff", "activity", "live", "cwd", "env",
	}
	if len(expectedGroups) != 15 {
		t.Fatalf("expected 15 groups in test list, got %d", len(expectedGroups))
	}

	modes := []segment.StyleMode{segment.StylePlain, segment.StyleNerdFont, segment.StylePowerline}
	for _, grp := range expectedGroups {
		for _, mode := range modes {
			m, ok := icons[grp]
			if !ok {
				t.Errorf("group %q missing from icons map", grp)
				continue
			}
			if _, ok := m[mode]; !ok {
				t.Errorf("group %q missing mode %q", grp, mode)
			}
		}
	}
}

func TestDefaultSeparator(t *testing.T) {
	tests := []struct {
		mode segment.StyleMode
		want string
	}{
		{segment.StylePlain, " | "},
		{segment.StyleNerdFont, " \u2502 "},
		{segment.StylePowerline, ""},
	}
	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			got := DefaultSeparator(tt.mode)
			if got != tt.want {
				t.Errorf("DefaultSeparator(%q) = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}

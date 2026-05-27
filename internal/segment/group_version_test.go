package segment

import (
	"testing"

	"github.com/ming-claude/ccsl/internal/input"
	"github.com/ming-claude/ccsl/internal/npm"
)

func TestStripCommonVersionPrefix(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    string
	}{
		{"patch bump", "2.1.142", "2.1.144", "144"},
		{"build segment bump", "2.1.142", "2.1.142.1", "1"},
		{"minor bump keeps full", "2.1.144", "2.2.0", "2.2.0"},
		{"major bump keeps full", "1.5.0", "2.0.0", "2.0.0"},
		{"empty current", "", "2.1.144", "2.1.144"},
		{"identical inputs return full latest", "2.1.144", "2.1.144", "2.1.144"},
		{"current shorter than latest", "2.1", "2.1.144", "144"},
		{"latest shorter than current", "2.1.144", "2.1", "2.1"},
		{"single-segment match", "2", "2.1.144", "2.1.144"},
		{"prerelease tail diff", "2.1.142-rc.1", "2.1.142-rc.2", "2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripCommonVersionPrefix(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("stripCommonVersionPrefix(%q, %q) = %q, want %q", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestVersionGroup_Render(t *testing.T) {
	tests := []struct {
		name   string
		ctx    *RenderContext
		want   string
		wantOk bool
	}{
		{
			name: "no npm info",
			ctx:  &RenderContext{Style: StylePlain},
		},
		{
			name: "no update",
			ctx: &RenderContext{
				Style:      StylePlain,
				NpmVersion: &npm.VersionInfo{Latest: "2.1.144", HasUpdate: false, Channel: "stable"},
			},
		},
		{
			name: "patch bump compresses",
			ctx: &RenderContext{
				Style:      StylePlain,
				NpmVersion: &npm.VersionInfo{Latest: "2.1.144", HasUpdate: true, Channel: "stable"},
				Stdin:      &input.StdinData{Version: "2.1.142"},
			},
			want:   "2.1.142 ⬆ 144 (stable)",
			wantOk: true,
		},
		{
			name: "minor bump keeps full latest",
			ctx: &RenderContext{
				Style:      StylePlain,
				NpmVersion: &npm.VersionInfo{Latest: "2.2.0", HasUpdate: true, Channel: "stable"},
				Stdin:      &input.StdinData{Version: "2.1.144"},
			},
			want:   "2.1.144 ⬆ 2.2.0 (stable)",
			wantOk: true,
		},
		{
			name: "no current version shows latest only",
			ctx: &RenderContext{
				Style:      StylePlain,
				NpmVersion: &npm.VersionInfo{Latest: "2.1.144", HasUpdate: true, Channel: "stable"},
			},
			want:   "2.1.144 (stable)",
			wantOk: true,
		},
		{
			name: "blank channel defaults to latest",
			ctx: &RenderContext{
				Style:      StylePlain,
				NpmVersion: &npm.VersionInfo{Latest: "2.1.144", HasUpdate: true, Channel: ""},
				Stdin:      &input.StdinData{Version: "2.1.142"},
			},
			want:   "2.1.142 ⬆ 144 (latest)",
			wantOk: true,
		},
	}

	g := &VersionGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := g.Render(tt.ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.wantOk {
				if out != nil {
					t.Fatalf("expected nil output, got %+v", out)
				}
				return
			}
			if out == nil {
				t.Fatalf("expected non-nil output")
			}
			if out.Primary != tt.want {
				t.Errorf("Primary = %q, want %q", out.Primary, tt.want)
			}
		})
	}
}

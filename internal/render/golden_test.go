package render

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestGolden(t *testing.T) {
	tests := []struct {
		name     string
		elements []RenderedElement
		sep      string
		width    int
	}{
		{
			name: "standard_3line_line1",
			elements: []RenderedElement{
				{Icon: "> ", Primary: "Opus 4.6", Priority: 10},
				{Icon: "v", Primary: "2.1.90", Priority: 4},
				{Icon: "~", Primary: "concise", Priority: 4},
			},
			sep:   " | ",
			width: 80,
		},
		{
			name: "standard_3line_line2",
			elements: []RenderedElement{
				{Primary: "████████░░", SeparatorAfter: " ", Priority: 10},
				{Icon: "Ctx:", Primary: "78%", Priority: 10},
				{Icon: "I:", Primary: "12.3k", SeparatorAfter: "/", Priority: 6},
				{Icon: "O:", Primary: "4.5k", SeparatorAfter: " ", Priority: 6},
				{Icon: "$", Primary: "0.42", Priority: 8},
			},
			sep:   " | ",
			width: 80,
		},
		{
			name: "narrow_terminal_truncation",
			elements: []RenderedElement{
				{Icon: "> ", Primary: "Opus 4.6", Priority: 10},
				{Primary: "/very/long/path/to/some/directory", Priority: 2, IsVariable: true, MinWidth: 8},
				{Icon: "*", Primary: "feature-branch-name", Priority: 8},
			},
			sep:   " | ",
			width: 40,
		},
		{
			name: "narrow_terminal_hide_low_priority",
			elements: []RenderedElement{
				{Icon: "> ", Primary: "Opus 4.6", Priority: 10},
				{Icon: "v", Primary: "2.1.90", Priority: 2},
				{Icon: "$", Primary: "0.42", Priority: 8},
			},
			sep:   " | ",
			width: 25,
		},
		{
			name: "colored_elements",
			elements: []RenderedElement{
				{Icon: "> ", Primary: "Opus 4.6", Color: "#61afef", Priority: 10},
				{Icon: "*", Primary: "main", Color: "#98c379", Priority: 8},
			},
			sep:   " │ ",
			width: 80,
		},
		{
			name: "single_element",
			elements: []RenderedElement{
				{Icon: "> ", Primary: "Opus 4.6", Priority: 10},
			},
			sep:   " | ",
			width: 80,
		},
		{
			name:     "empty_elements",
			elements: nil,
			sep:      " | ",
			width:    80,
		},
		{
			name: "block_colored_elements",
			elements: []RenderedElement{
				{Icon: "> ", Primary: "Opus 4.6", Color: "#282c34", BgColor: "#c678dd", Priority: 10},
				{Icon: "*", Primary: "main", Color: "#282c34", BgColor: "#98c379", Priority: 8},
			},
			sep:   " ",
			width: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderLine(tt.elements, tt.sep, tt.width)
			golden := filepath.Join("testdata", tt.name+".golden")

			if *update {
				if err := os.MkdirAll("testdata", 0755); err != nil {
					t.Fatalf("failed to create testdata dir: %v", err)
				}
				if err := os.WriteFile(golden, []byte(got), 0644); err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				t.Logf("updated golden file: %s", golden)
				return
			}

			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("missing golden file (run with -update flag): %v", err)
			}
			if got != string(want) {
				t.Errorf("output mismatch for %s\ngot:  %q\nwant: %q", tt.name, got, string(want))
			}
		})
	}
}

func TestGolden_MultiLine(t *testing.T) {
	tests := []struct {
		name  string
		lines [][]RenderedElement
		sep   string
		width int
	}{
		{
			name: "multiline_block_aligned",
			lines: [][]RenderedElement{
				{
					{Primary: "Opus 4.6", Color: "#282c34", BgColor: "#c678dd", Priority: 10},
					{Primary: "~/project", Color: "#282c34", BgColor: "#3c404d", Priority: 6},
					{Primary: "main", Color: "#282c34", BgColor: "#98c379", Priority: 8},
				},
				{
					{Primary: "78%", Color: "#282c34", BgColor: "#283d60", Priority: 10},
					{Primary: "12.3k", Color: "#282c34", BgColor: "#1e4545", Priority: 6},
					{Primary: "$0.42", Color: "#282c34", BgColor: "#4a3818", Priority: 8},
				},
				{
					{Primary: "5h: 82%", Color: "#282c34", BgColor: "#4a3020", Priority: 6},
					{Primary: "Wk: $4.20", Color: "#282c34", BgColor: "#1a3048", Priority: 6},
					{Primary: "v2.1", Color: "#282c34", BgColor: "#2e3038", Priority: 4},
				},
			},
			sep:   " ",
			width: 0,
		},
		{
			name: "multiline_block_narrow",
			lines: [][]RenderedElement{
				{
					{Primary: "Opus 4.6", Color: "#282c34", BgColor: "#c678dd", Priority: 10},
					{Primary: "~/my/long/project/path", Color: "#282c34", BgColor: "#3c404d", Priority: 6,
						IsVariable: true, MinWidth: 10},
					{Primary: "main", Color: "#282c34", BgColor: "#98c379", Priority: 8},
				},
				{
					{Primary: "78%", Color: "#282c34", BgColor: "#283d60", Priority: 10},
					{Primary: "12.3k", Color: "#282c34", BgColor: "#1e4545", Priority: 6},
					{Primary: "$0.42", Color: "#282c34", BgColor: "#4a3818", Priority: 8},
				},
			},
			sep:   " ",
			width: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := RenderAllLines(tt.lines, tt.sep, tt.width)
			got := ""
			for i, line := range result {
				if i > 0 {
					got += "\n"
				}
				got += line
			}
			golden := filepath.Join("testdata", tt.name+".golden")

			if *update {
				if err := os.MkdirAll("testdata", 0755); err != nil {
					t.Fatalf("failed to create testdata dir: %v", err)
				}
				if err := os.WriteFile(golden, []byte(got), 0644); err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				t.Logf("updated golden file: %s", golden)
				return
			}

			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("missing golden file (run with -update flag): %v", err)
			}
			if got != string(want) {
				t.Errorf("output mismatch for %s\ngot:  %q\nwant: %q", tt.name, got, string(want))
			}
		})
	}
}

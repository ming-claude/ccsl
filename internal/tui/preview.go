package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ming-claude/ccsl/internal/config"
	gitpkg "github.com/ming-claude/ccsl/internal/git"
	"github.com/ming-claude/ccsl/internal/input"
	"github.com/ming-claude/ccsl/internal/npm"
	"github.com/ming-claude/ccsl/internal/pipeline"
	"github.com/ming-claude/ccsl/internal/render"
	"github.com/ming-claude/ccsl/internal/segment"
	"github.com/ming-claude/ccsl/internal/style"
	"github.com/ming-claude/ccsl/internal/transcript"
)

// mockRenderContext builds a RenderContext with representative mock data
// so the preview uses the real segment Render() pipeline.
func mockRenderContext() *segment.RenderContext {
	return &segment.RenderContext{
		Stdin: &input.StdinData{
			CWD:         "/home/user/projects/my-app",
			SessionID:   "abc12345-6789-0def-ghij-klmnopqrstuv",
			SessionName: "my-project",
			Version:     "1.0.37",
			Model: &input.Model{
				DisplayName: "Claude Opus 4.6",
			},
			ContextWindow: &input.ContextWindow{
				TotalInputTokens:  52000,
				TotalOutputTokens: 8000,
				ContextWindowSize: 1000000,
				UsedPercentage:    42,
				CurrentUsage: &input.CurrentUsage{
					InputTokens:              18000,
					OutputTokens:             4000,
					CacheCreationInputTokens: 12000,
					CacheReadInputTokens:     22000,
				},
			},
			Cost: &input.Cost{
				TotalCostUSD:       0.42,
				TotalDurationMs:    872000,
				TotalAPIDurationMs: 45000,
				TotalLinesAdded:    87,
				TotalLinesRemoved:  23,
			},
			RateLimits: &input.RateLimits{
				FiveHour: &input.RateLimit{
					UsedPercentage: 35,
				},
				SevenDay: &input.RateLimit{
					UsedPercentage: 18,
				},
			},
			Vim:         &input.VimInfo{Mode: "NORMAL"},
			OutputStyle: &input.OutputStyle{Name: "normal"},
		},
		Git: &gitpkg.GitData{
			Branch:     "feat/ccsl",
			Ahead:      2,
			Behind:     0,
			HasRemote:  true,
			Modified:   3,
			Untracked:  1,
			Insertions: 142,
			Deletions:  38,
		},
		Transcript: &transcript.TranscriptData{
			ActiveTools: []transcript.ToolInfo{
				{Name: "Bash", Context: "go test ./..."},
			},
			CompletedTools: 14,
			ToolStats:      map[string]int{"Bash": 8, "Read": 4, "Edit": 2},
			ActiveAgents:   nil,
			AgentStats:     3,
			Todos:          []transcript.TodoItem{{Subject: "Fix preview", Status: "completed"}},
			TodoCompleted:  3,
			Skills:         []string{"commit", "plan"},
			UsageStats: transcript.UsageStats{
				InputTokens:              48000,
				OutputTokens:             5790,
				CacheCreationInputTokens: 320000,
				CacheReadInputTokens:     361000,
			},
		},
		NpmVersion: &npm.VersionInfo{
			Latest:     "1.0.38",
			HasUpdate:  true,
			CurrentVer: "1.0.37",
			Channel:    "latest",
		},
		MemoryPct: "67%",
		Style:     segment.StyleNerdFont,
	}
}

// buildPreview builds statusline output using the real segment pipeline
// with mock data, ensuring preview matches actual rendering.
func buildPreview(cfg *config.Config, width int) []string {
	styleMode := segment.StyleMode(cfg.Style)
	if styleMode == "" {
		styleMode = segment.StyleNerdFont
	}

	sep := cfg.Separator
	if !cfg.SeparatorSet && sep == "" {
		sep = style.DefaultSeparator(styleMode)
	}

	registry := segment.DefaultRegistry()
	ctx := mockRenderContext()
	ctx.Style = styleMode
	ctx.Disabled = cfg.Disabled

	allElements := pipeline.BuildElements(cfg.Lines, registry, ctx, cfg.ThemeColors, styleMode)

	maxWidth := width
	if maxWidth <= 0 {
		maxWidth = 0
	}
	lines, _ := render.RenderAllLines(allElements, sep, maxWidth)
	return lines
}

// previewBox renders the preview inside a styled border box.
func previewBox(cfg *config.Config, width, maxHeight int) string {
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	previewWidth := width - 6
	if previewWidth < 20 {
		previewWidth = 20
	}

	lines := buildPreview(cfg, previewWidth)
	if len(lines) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
		return dimStyle.Render("  No lines to preview.")
	}

	// Respect maxHeight.
	if maxHeight > 0 && len(lines) > maxHeight {
		lines = lines[:maxHeight]
	}

	content := strings.Join(lines, "\n")
	return borderStyle.Width(previewWidth).Render(content)
}

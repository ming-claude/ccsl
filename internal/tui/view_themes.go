package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ming-claude/ccsl/internal/config"
)

type themeItem struct {
	name    string
	variant string   // "block" or "non-block"
	colors  []string // representative colors for swatches
}

type themesView struct {
	entries []themeItem
	cursor  int
	active  string
	userCfg *config.UserConfig
}

func newThemesView(cfg *config.Config, userCfg *config.UserConfig) *themesView {
	names := config.ListThemes()
	var entries []themeItem
	for _, name := range names {
		theme, err := config.LoadTheme(name)
		if err != nil {
			continue
		}
		variant := theme.Variant
		if variant == "" {
			variant = "non-block"
		}
		// Pick representative colors for swatches using group names.
		var swatchColors []string
		reps := []string{"model", "context", "tokens", "git", "usage_5hour"}
		for _, group := range reps {
			if c, ok := theme.Colors[group]; ok {
				if variant == "block" && c.Bg != "" {
					swatchColors = append(swatchColors, c.Bg)
				} else {
					swatchColors = append(swatchColors, c.Fg)
				}
			}
		}
		entries = append(entries, themeItem{
			name:    name,
			variant: variant,
			colors:  swatchColors,
		})
	}

	return &themesView{
		entries: entries,
		cursor:  0,
		active:  cfg.Theme,
		userCfg: userCfg,
	}
}

func (v *themesView) Title() string { return "Select Theme" }

func (v *themesView) HelpKeys() string {
	return "j/k: navigate | Enter: apply theme | Esc: back | q: quit"
}

func (v *themesView) Update(msg tea.Msg, cfg *config.Config) (View, tea.Cmd, ViewAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if v.cursor < len(v.entries)-1 {
				v.cursor++
			}
		case "k", "up":
			if v.cursor > 0 {
				v.cursor--
			}
		case "enter":
			if v.cursor >= len(v.entries) {
				return v, nil, ViewNone
			}
			entry := v.entries[v.cursor]
			theme, err := config.LoadTheme(entry.name)
			if err != nil {
				return v, nil, ViewNone
			}
			cfg.Theme = entry.name
			cfg.ThemeColors = make(map[string]config.SegmentColor, len(theme.Colors))
			for k, c := range theme.Colors {
				cfg.ThemeColors[k] = c
			}
			if theme.Separator != nil {
				cfg.Separator = *theme.Separator
				cfg.SeparatorSet = true
			}
			// Sync userCfg.
			v.userCfg.Theme = entry.name
			v.active = entry.name
			return v, nil, ViewDirty
		case "esc":
			return v, nil, ViewPop
		case "q", "Q":
			return v, nil, ViewQuit
		}
	}
	return v, nil, ViewNone
}

func (v *themesView) View(ctx ViewContext) string {
	var b strings.Builder

	// Preview with currently highlighted theme.
	if v.cursor < len(v.entries) {
		entry := v.entries[v.cursor]
		previewCfg := buildThemePreviewConfig(entry.name)
		if previewCfg != nil {
			b.WriteString(previewBox(previewCfg, ctx.Width, 6))
		}
	}
	b.WriteString("\n\n")

	// Theme list.
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	variantStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	for i, entry := range v.entries {
		cursor := "  "
		if i == v.cursor {
			cursor = cursorStyle.Render("\u25b6 ")
		}
		active := ""
		if entry.name == v.active {
			active = activeStyle.Render(" (active)")
		}

		// Color swatches.
		var swatches strings.Builder
		for _, c := range entry.colors {
			swatches.WriteString(lipgloss.NewStyle().Background(lipgloss.Color(c)).Render("  "))
			swatches.WriteByte(' ')
		}

		fmt.Fprintf(&b, "  %s%s %s  %s%s\n",
			cursor,
			labelStyle.Render(entry.name),
			variantStyle.Render(entry.variant),
			swatches.String(),
			active,
		)
	}

	return b.String()
}

// buildThemePreviewConfig builds a temporary Config for previewing a theme.
func buildThemePreviewConfig(themeName string) *config.Config {
	cfg, err := config.Resolve(&config.UserConfig{Theme: themeName})
	if err != nil {
		return nil
	}
	return cfg
}

package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ming-claude/ccsl/internal/config"
)

type presetItem struct {
	name string
	desc string
}

var presetItems = []presetItem{
	{name: "minimal", desc: "Single line: model, context, cost, git"},
	{name: "standard", desc: "Three lines: essentials"},
	{name: "full", desc: "Five lines: everything"},
	{name: "dev", desc: "Four lines: dev-focused"},
}

type presetsView struct {
	cursor  int
	active  string
	userCfg *config.UserConfig
}

func newPresetsView(cfg *config.Config, userCfg *config.UserConfig) *presetsView {
	return &presetsView{
		cursor:  0,
		active:  cfg.Preset,
		userCfg: userCfg,
	}
}

func (v *presetsView) Title() string { return "Use Presets" }

func (v *presetsView) HelpKeys() string {
	return "j/k: navigate | Enter: apply preset | Esc: back | q: quit"
}

func (v *presetsView) Update(msg tea.Msg, cfg *config.Config) (View, tea.Cmd, ViewAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if v.cursor < len(presetItems)-1 {
				v.cursor++
			}
		case "k", "up":
			if v.cursor > 0 {
				v.cursor--
			}
		case "enter":
			name := presetItems[v.cursor].name
			preset, err := config.LoadPreset(name)
			if err != nil {
				return v, nil, ViewNone
			}
			// Apply preset: overwrite lines and disabled.
			cfg.Preset = name
			cfg.Lines = preset.Lines
			// Rebuild disabled from preset defaults + user disabled_segments.
			cfg.Disabled = make(map[string]bool)
			for _, s := range preset.Disabled {
				cfg.Disabled[s] = true
			}
			for _, s := range v.userCfg.DisabledSegments {
				cfg.Disabled[s] = true
			}
			// Reload theme colors for the current theme.
			if theme, err := config.LoadTheme(cfg.Theme); err == nil {
				cfg.ThemeColors = make(map[string]config.SegmentColor, len(theme.Colors))
				for k, c := range theme.Colors {
					cfg.ThemeColors[k] = c
				}
				if theme.Separator != nil {
					cfg.Separator = *theme.Separator
					cfg.SeparatorSet = true
				}
			}
			// Sync userCfg.
			v.userCfg.Preset = name
			v.active = name
			return v, nil, ViewDirty
		case "esc":
			return v, nil, ViewPop
		case "q", "Q":
			return v, nil, ViewQuit
		}
	}
	return v, nil, ViewNone
}

func (v *presetsView) View(ctx ViewContext) string {
	var b strings.Builder

	// Preview (uses currently highlighted preset for preview).
	name := presetItems[v.cursor].name
	previewCfg := buildPresetPreviewConfig(name)
	if previewCfg != nil {
		b.WriteString(previewBox(previewCfg, ctx.Width, 6))
	}
	b.WriteString("\n\n")

	// Preset list.
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	for i, p := range presetItems {
		cursor := "  "
		if i == v.cursor {
			cursor = cursorStyle.Render("\u25b6 ")
		}
		active := ""
		if p.name == v.active {
			active = activeStyle.Render(" (active)")
		}
		fmt.Fprintf(&b, "  %s%s %s%s\n",
			cursor,
			labelStyle.Render(p.name),
			descStyle.Render("\u2014 "+p.desc),
			active,
		)
	}

	return b.String()
}

// buildPresetPreviewConfig builds a temporary Config for previewing a preset.
func buildPresetPreviewConfig(presetName string) *config.Config {
	cfg, err := config.Resolve(&config.UserConfig{Preset: presetName})
	if err != nil {
		return nil
	}
	return cfg
}

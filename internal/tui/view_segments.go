package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ming-claude/ccsl/internal/config"
)

// groupChildren maps each group to its toggleable child segments.
// Groups with no children (standalone) have an empty slice.
var groupChildren = map[string][]string{
	"model":        {"model"},
	"git":          {"branch", "changes", "insertions", "deletions", "worktree"},
	"context":      {"bar", "pct", "length"},
	"tokens":       {"input", "output", "cached", "total"},
	"cost":         {},
	"usage_5hour":  {},
	"usage_weekly": {},
	"version":      {},
	"session":      {"id", "name", "clock"},
	"speed":        {"input", "output", "total"},
	"diff":         {"added", "removed"},
	"activity":     {"tool_stats", "todo_progress", "skills"},
	"live":         {"tool_active", "agent_active"},
	"cwd":          {},
	"env":          {"vim_mode", "output_style", "terminal_width", "memory_usage"},
}

// segmentRow is a flat list item: either a group header or a child.
type segmentRow struct {
	group string // group name
	child string // empty for group rows
}

func (r segmentRow) key() string {
	if r.child == "" {
		return r.group
	}
	return r.group + "." + r.child
}

type segmentsView struct {
	rows    []segmentRow
	cursor  int
	userCfg *config.UserConfig
	scroll  int // top visible row for scrolling
}

func newSegmentsView(cfg *config.Config, userCfg *config.UserConfig) *segmentsView {
	v := &segmentsView{userCfg: userCfg}
	v.buildRows(cfg)
	return v
}

// buildRows rebuilds the flat row list from the current config's lines.
func (v *segmentsView) buildRows(cfg *config.Config) {
	v.rows = nil
	seen := make(map[string]bool)
	for _, line := range cfg.Lines {
		for _, group := range line {
			if seen[group] {
				continue
			}
			seen[group] = true
			v.rows = append(v.rows, segmentRow{group: group})
			for _, child := range groupChildren[group] {
				v.rows = append(v.rows, segmentRow{group: group, child: child})
			}
		}
	}
}

// isDisabled checks if a key is in userCfg.DisabledSegments.
func (v *segmentsView) isDisabled(key string) bool {
	for _, s := range v.userCfg.DisabledSegments {
		if s == key {
			return true
		}
	}
	return false
}

// toggle adds or removes a key from userCfg.DisabledSegments.
func (v *segmentsView) toggle(key string) {
	for i, s := range v.userCfg.DisabledSegments {
		if s == key {
			// Remove it.
			v.userCfg.DisabledSegments = append(
				v.userCfg.DisabledSegments[:i],
				v.userCfg.DisabledSegments[i+1:]...,
			)
			return
		}
	}
	// Not found — add it.
	v.userCfg.DisabledSegments = append(v.userCfg.DisabledSegments, key)
}

func (v *segmentsView) Title() string { return "Segments" }

func (v *segmentsView) HelpKeys() string {
	return "j/k: navigate | Enter/Space: toggle | Esc: back | q: quit"
}

func (v *segmentsView) Update(msg tea.Msg, cfg *config.Config) (View, tea.Cmd, ViewAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if v.cursor < len(v.rows)-1 {
				v.cursor++
			}
		case "k", "up":
			if v.cursor > 0 {
				v.cursor--
			}
		case "enter", " ":
			if v.cursor < len(v.rows) {
				row := v.rows[v.cursor]
				v.toggle(row.key())
				// Re-resolve cfg from userCfg.
				v.syncDisabled(cfg)
				return v, nil, ViewDirty
			}
		case "esc":
			return v, nil, ViewPop
		case "q", "Q":
			return v, nil, ViewQuit
		}
	}
	return v, nil, ViewNone
}

// syncDisabled rebuilds cfg.Disabled from the current preset defaults + userCfg.DisabledSegments.
func (v *segmentsView) syncDisabled(cfg *config.Config) {
	// Reload preset to get its default disabled list.
	preset, err := config.LoadPreset(cfg.Preset)
	if err != nil {
		return
	}
	disabled := make(map[string]bool)
	for _, s := range preset.Disabled {
		disabled[s] = true
	}
	for _, s := range v.userCfg.DisabledSegments {
		disabled[s] = true
	}
	cfg.Disabled = disabled
}

func (v *segmentsView) View(ctx ViewContext) string {
	var b strings.Builder

	// Preview box at the top — build a temporary config for preview.
	previewCfg := v.buildPreviewConfig()
	if previewCfg != nil {
		b.WriteString(previewBox(previewCfg, ctx.Width, 6))
	}
	b.WriteString("\n\n")

	// Styles.
	groupStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	childStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	disabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	checkOn := lipgloss.NewStyle().Foreground(lipgloss.Color("114")).Render("\u2713")
	checkOff := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("\u2717")
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	// Calculate visible area for scrolling.
	// Reserve lines already used: preview (~8 lines) + 2 newlines.
	// We'll allow the list to use remaining height.
	listHeight := ctx.Height - 10
	if listHeight < 5 {
		listHeight = 5
	}

	// Adjust scroll to keep cursor visible.
	if v.cursor < v.scroll {
		v.scroll = v.cursor
	}
	if v.cursor >= v.scroll+listHeight {
		v.scroll = v.cursor - listHeight + 1
	}

	end := v.scroll + listHeight
	if end > len(v.rows) {
		end = len(v.rows)
	}

	for i := v.scroll; i < end; i++ {
		row := v.rows[i]
		key := row.key()
		disabled := v.isDisabled(key)

		cursor := "  "
		if i == v.cursor {
			cursor = cursorStyle.Render("\u25b6 ")
		}

		check := checkOn
		if disabled {
			check = checkOff
		}

		if row.child == "" {
			// Group row.
			label := groupStyle.Render(row.group)
			if disabled {
				label = disabledStyle.Render(row.group)
			}
			fmt.Fprintf(&b, "  %s%s %s\n", cursor, check, label)
		} else {
			// Child row (indented).
			label := childStyle.Render(row.child)
			if disabled {
				label = disabledStyle.Render(row.child)
			}
			fmt.Fprintf(&b, "  %s  %s %s\n", cursor, check, label)
		}
	}

	return b.String()
}

// buildPreviewConfig creates a Config for the preview using current userCfg state.
func (v *segmentsView) buildPreviewConfig() *config.Config {
	cfg, err := config.Resolve(v.userCfg)
	if err != nil {
		return nil
	}
	return cfg
}

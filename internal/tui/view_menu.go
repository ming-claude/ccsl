package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ming-claude/ccsl/internal/config"
)

type menuItem struct {
	label string
	desc  string
	badge string
}

type menuView struct {
	items    []menuItem
	cursor   int
	selected int // index of selected item, or -1
}

func newMenuView(cfg *config.Config) *menuView {
	m := &menuView{
		cursor:   0,
		selected: -1,
	}
	m.refresh(cfg)
	return m
}

func (m *menuView) refresh(cfg *config.Config) {
	// Detect install status.
	installBadge := "not installed"
	if isInstalled() {
		installBadge = "installed"
	}

	// Count disabled segments for badge.
	disabledCount := len(cfg.Disabled)
	disabledBadge := ""
	if disabledCount > 0 {
		disabledBadge = fmt.Sprintf("%d disabled", disabledCount)
	}

	debugBadge := "off"
	if cfg.Debug {
		debugBadge = "on"
	}

	m.items = []menuItem{
		{label: "Use Presets", desc: "Apply a preset layout", badge: cfg.Preset},
		{label: "Select Theme", desc: "Switch color theme", badge: cfg.Theme},
		{label: "Segments", desc: "Toggle groups and segments", badge: disabledBadge},
		{label: "Debug Logging", desc: "Write diagnostic data to debug.jsonl", badge: debugBadge},
		{label: "Install to Claude", desc: "Set up ~/.claude/settings.json", badge: installBadge},
	}
}

func (m *menuView) Title() string {
	return "ccsl config"
}

func (m *menuView) HelpKeys() string {
	return "j/k: navigate | Enter: select | S: save | q/Esc: quit"
}

func (m *menuView) Update(msg tea.Msg, _ *config.Config) (View, tea.Cmd, ViewAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			m.selected = m.cursor
		case "q", "Q", "esc":
			return m, nil, ViewQuit
		}
	}
	return m, nil, ViewNone
}

func (m *menuView) View(ctx ViewContext) string {
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	badgeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	var b strings.Builder
	b.WriteByte('\n')
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("\u25b6 ")
		}
		badge := ""
		if item.badge != "" {
			badge = "  " + badgeStyle.Render("["+item.badge+"]")
		}
		fmt.Fprintf(&b, "  %s%s %s%s\n",
			cursor,
			labelStyle.Render(item.label),
			descStyle.Render("\u2014 "+item.desc),
			badge,
		)
	}
	return b.String()
}

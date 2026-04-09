package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ming-claude/ccsl/internal/config"
)

// Model is the top-level bubbletea model for the TUI configurator.
type Model struct {
	cfg      *config.Config
	userCfg  *config.UserConfig
	dirty    bool
	width    int
	height   int
	quitting bool // confirm-quit state

	stack []View // navigation stack; stack[0] = main menu
}

// NewModel creates a new TUI model from an existing config.
func NewModel(cfg *config.Config, userCfg *config.UserConfig) *Model {
	m := &Model{
		cfg:     cfg,
		userCfg: userCfg,
		stack:   []View{newMenuView(cfg)},
	}
	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.WindowSize()
}

// activeView returns the top of the navigation stack.
func (m *Model) activeView() View {
	return m.stack[len(m.stack)-1]
}

// pushView pushes a new view onto the stack.
func (m *Model) pushView(v View) {
	m.stack = append(m.stack, v)
}

// popView removes the top view from the stack (never pops the menu).
func (m *Model) popView() {
	if len(m.stack) > 1 {
		m.stack = m.stack[:len(m.stack)-1]
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		key := msg.String()

		// Confirm-quit dialog.
		if m.quitting {
			switch key {
			case "y", "Y":
				return m, tea.Quit
			case "s", "S":
				if err := config.Save(m.userCfg); err == nil {
					m.dirty = false
				}
				return m, tea.Quit
			default:
				m.quitting = false
				return m, nil
			}
		}

		// Global keys (always active).
		switch key {
		case "ctrl+c":
			return m, tea.Quit

		case "S":
			if err := config.Save(m.userCfg); err == nil {
				m.dirty = false
			}
			return m, nil
		}
	}

	// Delegate to active view.
	view := m.activeView()
	newView, cmd, action := view.Update(msg, m.cfg)
	m.stack[len(m.stack)-1] = newView

	switch action {
	case ViewPop:
		m.popView()
		// Refresh menu badges when returning.
		if menu, ok := m.activeView().(*menuView); ok {
			menu.refresh(m.cfg)
		}
	case ViewDirty:
		m.dirty = true
	case ViewQuit:
		if m.dirty {
			m.quitting = true
		} else {
			return m, tea.Quit
		}
	}

	// Handle menu item selection - push new view.
	if menu, ok := m.activeView().(*menuView); ok {
		if menu.selected >= 0 {
			idx := menu.selected
			menu.selected = -1
			switch idx {
			case 0:
				m.pushView(newPresetsView(m.cfg, m.userCfg))
			case 1:
				m.pushView(newThemesView(m.cfg, m.userCfg))
			case 2:
				m.pushView(newSegmentsView(m.cfg, m.userCfg))
			case 3:
				// Debug toggle — flip in place, no sub-view.
				m.userCfg.Debug = !m.userCfg.Debug
				m.cfg.Debug = m.userCfg.Debug
				m.dirty = true
				menu.refresh(m.cfg)
			case 4:
				m.pushView(newInstallView())
			}
		}
	}

	return m, cmd
}

func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	ctx := ViewContext{
		Width:  m.width,
		Height: m.height,
	}

	// Title bar.
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("62")).
		Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	title := titleStyle.Render(m.activeView().Title())
	dirtyMark := ""
	if m.dirty {
		dirtyMark = dimStyle.Render(" [modified]")
	}
	titleBar := title + dirtyMark

	// Content area: leave room for title (1 line) + gap (1 line) + help (1 line) + gap (1 line).
	contentCtx := ViewContext{
		Width:  ctx.Width,
		Height: ctx.Height - 4,
	}
	if contentCtx.Height < 3 {
		contentCtx.Height = 3
	}
	content := m.activeView().View(contentCtx)

	// Help bar.
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	helpBar := helpStyle.Render(m.activeView().HelpKeys())

	page := fmt.Sprintf("%s\n\n%s\n\n%s", titleBar, content, helpBar)

	// Overlay modal dialog when quitting with unsaved changes.
	if m.quitting {
		dialog := m.renderQuitDialog()
		page = m.overlayCenter(page, dialog)
	}

	return page
}

// renderQuitDialog renders a modal dialog box for unsaved-changes confirmation.
func (m *Model) renderQuitDialog() string {
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(1, 3)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("114"))

	content := fmt.Sprintf(
		"%s\n\n%s\n\n  %s  Save and quit\n  %s  Quit without saving\n  %s  Cancel",
		titleStyle.Render("Unsaved Changes"),
		textStyle.Render("You have unsaved changes."),
		keyStyle.Render("s"),
		keyStyle.Render("y"),
		keyStyle.Render("n"),
	)

	return borderStyle.Render(content)
}

// overlayCenter places a dialog box centered over the background content.
func (m *Model) overlayCenter(bg, dialog string) string {
	bgLines := strings.Split(bg, "\n")
	dialogLines := strings.Split(dialog, "\n")

	// Pad background to fill terminal height.
	for len(bgLines) < m.height {
		bgLines = append(bgLines, "")
	}

	// Calculate centered position.
	dialogHeight := len(dialogLines)
	dialogWidth := lipgloss.Width(dialog)
	startRow := (m.height - dialogHeight) / 2
	startCol := (m.width - dialogWidth) / 2
	if startRow < 0 {
		startRow = 0
	}
	if startCol < 0 {
		startCol = 0
	}

	// Overlay dialog lines onto background.
	for i, dLine := range dialogLines {
		row := startRow + i
		if row >= len(bgLines) {
			break
		}
		// Build the line: padding + dialog content.
		pad := strings.Repeat(" ", startCol)
		bgLines[row] = pad + dLine
	}

	return strings.Join(bgLines[:m.height], "\n")
}

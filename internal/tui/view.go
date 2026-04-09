package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ming-claude/ccsl/internal/config"
)

// ViewContext provides terminal dimensions to views for rendering.
type ViewContext struct {
	Width  int
	Height int
}

// ViewAction signals the navigation stack how to respond after a view update.
type ViewAction int

const (
	ViewNone  ViewAction = iota
	ViewPop              // Esc pressed - pop this view from the stack
	ViewDirty            // config was modified - mark dirty
	ViewQuit             // quit the application
)

// View is the interface all TUI views implement.
type View interface {
	Update(msg tea.Msg, cfg *config.Config) (View, tea.Cmd, ViewAction)
	View(ctx ViewContext) string
	// Title returns the view's title for the title bar.
	Title() string
	// HelpKeys returns the help text for the bottom bar.
	HelpKeys() string
}

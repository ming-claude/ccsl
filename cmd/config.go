package cmd

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ming-claude/ccsl/internal/config"
	"github.com/ming-claude/ccsl/internal/tui"
)

// RunConfig launches the TUI configuration interface.
func RunConfig() error {
	cfg, userCfg, err := config.LoadWithUserConfig()
	if err != nil {
		return err
	}
	m := tui.NewModel(cfg, userCfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

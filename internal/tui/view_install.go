package tui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gojson "github.com/goccy/go-json"

	"github.com/ming-claude/ccsl/internal/config"
)

type installView struct {
	binaryPath string
	targetFile string
	command    string
	installed  bool
	err        error
	done       bool
}

func newInstallView() *installView {
	// Detect binary path.
	binPath := "ccsl"
	if p, err := exec.LookPath("ccsl"); err == nil {
		binPath = p
	}

	home, _ := os.UserHomeDir()
	target := filepath.Join(home, ".claude", "settings.json")

	// Check if already installed.
	installed := false
	if data, err := os.ReadFile(target); err == nil {
		var settings map[string]any
		if gojson.Unmarshal(data, &settings) == nil {
			if sl, ok := settings["statusLine"].(map[string]any); ok {
				if cmd, ok := sl["command"].(string); ok && filepath.Base(cmd) == "ccsl" {
					installed = true
				}
			}
		}
	}

	return &installView{
		binaryPath: binPath,
		targetFile: target,
		command:    "ccsl",
		installed:  installed,
	}
}

func (v *installView) Title() string { return "Install to Claude" }

func (v *installView) HelpKeys() string {
	if v.installed || v.done {
		return "Esc: back | q: quit"
	}
	return "Enter: install | Esc: back | q: quit"
}

func (v *installView) Update(msg tea.Msg, _ *config.Config) (View, tea.Cmd, ViewAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if !v.installed && !v.done {
				v.err = v.doInstall()
				v.done = true
				if v.err == nil {
					v.installed = true
				}
			}
		case "esc":
			return v, nil, ViewPop
		case "q", "Q":
			return v, nil, ViewQuit
		}
	}
	return v, nil, ViewNone
}

func (v *installView) doInstall() error {
	// Read existing settings.json.
	var settings map[string]any
	data, err := os.ReadFile(v.targetFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("read settings: %w", err)
		}
		settings = make(map[string]any)
	} else {
		if err := gojson.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parse settings: %w", err)
		}
	}

	// Set the statusLine config.
	settings["statusLine"] = map[string]any{
		"type":    "command",
		"command": v.command,
	}

	// Atomic write.
	newData, err := gojson.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	newData = append(newData, '\n')

	dir := filepath.Dir(v.targetFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	tmp := v.targetFile + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	if _, err := f.Write(newData); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("sync: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close temp: %w", err)
	}
	return os.Rename(tmp, v.targetFile)
}

func (v *installView) View(_ ViewContext) string {
	var b strings.Builder
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	fmt.Fprintf(&b, "\n  %s  %s\n", dimStyle.Render("Binary path:"), valStyle.Render(v.binaryPath))
	fmt.Fprintf(&b, "  %s  %s\n", dimStyle.Render("Target file:"), valStyle.Render(v.targetFile))
	fmt.Fprintf(&b, "  %s      %s\n", dimStyle.Render("Command:"), valStyle.Render(v.command))
	b.WriteByte('\n')

	if v.installed && v.done {
		fmt.Fprintf(&b, "  %s\n", successStyle.Render("\u2713 Installed successfully"))
		fmt.Fprintf(&b, "  %s\n", dimStyle.Render(`Added "statusLine": {"type":"command","command":"ccsl"} to settings.json`))
	} else if v.installed {
		fmt.Fprintf(&b, "  %s\n", successStyle.Render("\u2713 Already installed"))
	} else if v.done && v.err != nil {
		fmt.Fprintf(&b, "  %s %s\n", errStyle.Render("\u2717 Error:"), errStyle.Render(v.err.Error()))
	} else {
		fmt.Fprintf(&b, "  %s\n", dimStyle.Render("Press Enter to install."))
	}

	return b.String()
}

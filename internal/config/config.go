package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	gojson "github.com/goccy/go-json"
	"github.com/ming-claude/ccsl/presets"
	"github.com/ming-claude/ccsl/themes"
)

const (
	defaultPreset  = "standard"
	defaultTheme   = "catppuccin-block"
	defaultStyle   = "nerd-font"
	configDirName  = "ccsl"
	configFileName = "config.json"
)

// PresetConfig represents the parsed content of a preset JSON file.
type PresetConfig struct {
	Lines    [][]string `json:"lines"`
	Disabled []string   `json:"disabled"`
}

// ThemeConfig represents a theme file's content.
type ThemeConfig struct {
	Variant   string                  `json:"variant"`
	Separator *string                 `json:"separator"`
	Colors    map[string]SegmentColor `json:"colors"`
}

// LoadTheme loads a theme by name from the embedded FS.
func LoadTheme(name string) (*ThemeConfig, error) {
	filename := name + ".json"
	data, err := themes.FS.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("load theme %q: %w", name, err)
	}

	var theme ThemeConfig
	if err := gojson.Unmarshal(data, &theme); err != nil {
		return nil, fmt.Errorf("load theme %q: %w", name, err)
	}
	return &theme, nil
}

// ListThemes returns the names of all available built-in themes.
func ListThemes() []string {
	entries, err := themes.FS.ReadDir(".")
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".json") {
			names = append(names, strings.TrimSuffix(name, ".json"))
		}
	}
	return names
}

// LoadPreset loads a preset by name from the embedded FS.
// Returns lines (group name arrays) and a default disabled list.
func LoadPreset(name string) (*PresetConfig, error) {
	filename := name + ".json"
	data, err := presets.FS.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("load preset %q: %w", name, err)
	}

	var pc PresetConfig
	if err := gojson.Unmarshal(data, &pc); err != nil {
		return nil, fmt.Errorf("load preset %q: %w", name, err)
	}
	return &pc, nil
}

// configDir returns the path to the CCSL config directory (~/.claude/ccsl).
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".claude", configDirName), nil
}

// configPath returns the full path to the user config file.
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// loadUserConfig loads the user configuration from ~/.claude/ccsl/config.json.
// Returns nil, nil if the file does not exist.
func loadUserConfig() (*UserConfig, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read user config: %w", err)
	}

	var uc UserConfig
	if err := gojson.Unmarshal(data, &uc); err != nil {
		return nil, fmt.Errorf("parse user config: %w", err)
	}
	return &uc, nil
}

// Load loads the final resolved runtime configuration.
// Steps: read UserConfig -> load preset -> load theme -> merge disabled -> resolve defaults.
func Load() (*Config, error) {
	uc, err := loadUserConfig()
	if err != nil {
		return nil, err
	}
	return Resolve(uc)
}

// LoadWithUserConfig loads both the resolved Config and the raw UserConfig.
// The UserConfig is needed by the TUI for saving.
func LoadWithUserConfig() (*Config, *UserConfig, error) {
	uc, err := loadUserConfig()
	if err != nil {
		return nil, nil, err
	}
	if uc == nil {
		uc = &UserConfig{}
	}
	cfg, err := Resolve(uc)
	if err != nil {
		return nil, nil, err
	}
	return cfg, uc, nil
}

// Resolve builds a runtime Config from a UserConfig.
// If uc is nil, default values are used.
func Resolve(uc *UserConfig) (*Config, error) {
	if uc == nil {
		uc = &UserConfig{}
	}

	// Determine preset, theme, style with defaults.
	presetName := uc.Preset
	if presetName == "" {
		presetName = defaultPreset
	}
	themeName := uc.Theme
	if themeName == "" {
		themeName = defaultTheme
	}
	style := uc.Style
	if style == "" {
		style = defaultStyle
	}

	// Load preset.
	preset, err := LoadPreset(presetName)
	if err != nil {
		return nil, err
	}

	// Load theme with fallback to default if configured theme is missing.
	theme, err := LoadTheme(themeName)
	if err != nil {
		slog.Warn("configured theme not found, falling back to default",
			"theme", themeName, "default", defaultTheme)
		theme, err = LoadTheme(defaultTheme)
		if err != nil {
			return nil, fmt.Errorf("theme %q not found and fallback %q also failed: %w",
				themeName, defaultTheme, err)
		}
		themeName = defaultTheme
	}

	// Auto-downgrade block theme when truecolor not available.
	// Claude Code's Ink TUI re-renders ANSI through chalk, which
	// downgrades RGB backgrounds based on supports-color detection.
	// Switching to non-block avoids unusable black backgrounds.
	colorLevel := DetectColorLevel()
	if colorLevel < ColorLevelTruecolor && theme.Variant == "block" {
		nonBlockName := strings.TrimSuffix(themeName, "-block")
		if nonBlockName != themeName {
			if nbTheme, err := LoadTheme(nonBlockName); err == nil {
				slog.Info("truecolor not detected, downgrading to non-block theme",
					"from", themeName, "to", nonBlockName, "color_level", colorLevel)
				theme = nbTheme
				themeName = nonBlockName
			}
		}
	}

	// Build disabled set: union of preset defaults + user disabled_segments.
	disabled := make(map[string]bool)
	for _, s := range preset.Disabled {
		disabled[s] = true
	}
	for _, s := range uc.DisabledSegments {
		disabled[s] = true
	}

	// Resolve separator from theme.
	var separator string
	var separatorSet bool
	if theme.Separator != nil {
		separator = *theme.Separator
		separatorSet = true
	}

	return &Config{
		Preset:       presetName,
		Theme:        themeName,
		Style:        style,
		Lines:        preset.Lines,
		Disabled:     disabled,
		ThemeColors:  theme.Colors,
		Separator:    separator,
		SeparatorSet: separatorSet,
		Debug:        uc.Debug,
		Snapshot:     uc.Snapshot,
	}, nil
}

// Save writes the UserConfig to ~/.claude/ccsl/config.json.
// Uses atomic write (temp file + fsync + rename) for safety.
func Save(uc *UserConfig) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := gojson.MarshalIndent(uc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')

	path, err := configPath()
	if err != nil {
		return err
	}

	// Atomic write: temp file -> fsync -> rename.
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("sync temp config: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close temp config: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename temp config: %w", err)
	}

	return nil
}

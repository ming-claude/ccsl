package config

import "testing"

func TestDetectColorLevel(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		wantLevel int
	}{
		{
			name:      "COLORTERM=truecolor",
			env:       map[string]string{"COLORTERM": "truecolor"},
			wantLevel: 3,
		},
		{
			name:      "COLORTERM=24bit",
			env:       map[string]string{"COLORTERM": "24bit"},
			wantLevel: 3,
		},
		{
			name:      "TERM contains ghostty",
			env:       map[string]string{"TERM": "xterm-ghostty"},
			wantLevel: 3,
		},
		{
			name:      "TERM contains kitty",
			env:       map[string]string{"TERM": "xterm-kitty"},
			wantLevel: 3,
		},
		{
			name:      "TERM contains wezterm",
			env:       map[string]string{"TERM": "wezterm"},
			wantLevel: 3,
		},
		{
			name:      "TERM=xterm-256color",
			env:       map[string]string{"TERM": "xterm-256color"},
			wantLevel: 2,
		},
		{
			name:      "TERM=screen-256color",
			env:       map[string]string{"TERM": "screen-256color"},
			wantLevel: 2,
		},
		{
			name:      "TERM=xterm",
			env:       map[string]string{"TERM": "xterm"},
			wantLevel: 1,
		},
		{
			name:      "TERM=screen",
			env:       map[string]string{"TERM": "screen"},
			wantLevel: 1,
		},
		{
			name:      "TERM=linux",
			env:       map[string]string{"TERM": "linux"},
			wantLevel: 1,
		},
		{
			name:      "NO_COLOR set",
			env:       map[string]string{"NO_COLOR": "1", "COLORTERM": "truecolor"},
			wantLevel: 0,
		},
		{
			name:      "TERM=dumb",
			env:       map[string]string{"TERM": "dumb"},
			wantLevel: 0,
		},
		{
			name:      "FORCE_COLOR=3",
			env:       map[string]string{"FORCE_COLOR": "3", "TERM": "dumb"},
			wantLevel: 3,
		},
		{
			name:      "FORCE_COLOR=0",
			env:       map[string]string{"FORCE_COLOR": "0", "COLORTERM": "truecolor"},
			wantLevel: 0,
		},
		{
			name:      "FORCE_COLOR=1",
			env:       map[string]string{"FORCE_COLOR": "1"},
			wantLevel: 1,
		},
		{
			name:      "empty env defaults to 1",
			env:       map[string]string{},
			wantLevel: 1,
		},
		{
			name:      "256color in TERM beats basic xterm",
			env:       map[string]string{"TERM": "tmux-256color"},
			wantLevel: 2,
		},
		{
			name:      "COLORTERM truecolor beats TERM without 256color",
			env:       map[string]string{"TERM": "xterm", "COLORTERM": "truecolor"},
			wantLevel: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectColorLevelFromEnv(func(key string) string {
				return tt.env[key]
			})
			if got != tt.wantLevel {
				t.Errorf("got level %d, want %d", got, tt.wantLevel)
			}
		})
	}
}

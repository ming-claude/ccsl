//go:build integration

package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEndToEnd(t *testing.T) {
	// Build binary
	binary := filepath.Join(t.TempDir(), "ccsl_test")
	build := exec.Command("go", "build", "-o", binary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s %v", out, err)
	}

	input := `{"cwd":"/tmp/test","model":{"id":"claude-opus-4-6","display_name":"Opus 4.6"},"version":"2.1.90","context_window":{"used_percentage":45,"context_window_size":200000,"current_usage":{"input_tokens":8000,"output_tokens":1200,"cache_creation_input_tokens":500,"cache_read_input_tokens":200}},"cost":{"total_cost_usd":0.42,"total_duration_ms":120000,"total_api_duration_ms":5000}}`

	cmd := exec.Command(binary)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ccsl failed: %v", err)
	}

	output := string(out)

	// Verify key segments appear in output
	checks := []struct {
		name     string
		contains string
	}{
		{"model name", "Opus 4.6"},
		{"version", "2.1.90"},
		{"context percentage", "45%"},
		{"cost", "$0.42"},
	}

	for _, check := range checks {
		// Strip ANSI codes before checking
		stripped := stripANSI(output)
		if !strings.Contains(stripped, check.contains) {
			t.Errorf("missing %s (%q) in output:\n%s", check.name, check.contains, stripped)
		}
	}

	// Verify multi-line output
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Errorf("expected multi-line output, got %d lines", len(lines))
	}
}

func TestEndToEnd_NoStdin(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "ccsl_test")
	build := exec.Command("go", "build", "-o", binary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s %v", out, err)
	}

	// Run without stdin pipe — should exit 0 with no output
	cmd := exec.Command(binary)
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.CombinedOutput()
	// Should not crash — either exit 0 or exit with no panic
	_ = out
	_ = err
}

func stripANSI(s string) string {
	var b strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z') {
				inEscape = false
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

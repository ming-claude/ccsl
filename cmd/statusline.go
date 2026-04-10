package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ming-claude/ccsl/internal/config"
	"github.com/ming-claude/ccsl/internal/debug"
	gitpkg "github.com/ming-claude/ccsl/internal/git"
	"github.com/ming-claude/ccsl/internal/input"
	"github.com/ming-claude/ccsl/internal/npm"
	"github.com/ming-claude/ccsl/internal/pipeline"
	"github.com/ming-claude/ccsl/internal/render"
	"github.com/ming-claude/ccsl/internal/segment"
	"github.com/ming-claude/ccsl/internal/style"
	"github.com/ming-claude/ccsl/internal/transcript"
)

// RunStatusline is the main pipeline: read stdin, fetch data, render, output.
func RunStatusline() error {
	defer func() {
		if r := recover(); r != nil {
			// Silently recover — never crash Claude Code's statusline rendering.
			return
		}
	}()

	startTime := time.Now()
	fetchDetail := make(map[string]int64)

	// 1. Read stdin — nil means TTY (no pipe), exit silently.
	stdin, rawStdin, err := input.ReadStdin()
	if err != nil || stdin == nil {
		return nil
	}
	stdinDone := time.Now()

	// 2. Load config.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	configDone := time.Now()

	// Set debug mode from config.
	debug.SetEnabled(cfg.Debug)
	debugOn := debug.Enabled()

	// 3. Concurrent data fetch.
	// Per-goroutine timing captured in local vars to avoid concurrent map writes.
	var (
		gitData     *gitpkg.GitData
		transcData  *transcript.TranscriptData
		versionInfo *npm.VersionInfo
		memoryPct   string

		tGit, tTranscript, tNpm, tMemory int64
		gitTimingLocal                   map[string]int64
	)

	g := new(errgroup.Group)

	g.Go(func() error {
		t0 := time.Now()
		defer func() { tGit = time.Since(t0).Milliseconds() }()
		cwd := stdin.CWD
		if cwd == "" && stdin.Workspace != nil {
			cwd = stdin.Workspace.CurrentDir
		}
		if debugOn {
			gitTimingLocal = make(map[string]int64)
		}
		if cwd != "" {
			gitData, _ = gitpkg.GetData(cwd, gitTimingLocal)
		}
		return nil
	})

	g.Go(func() error {
		t0 := time.Now()
		defer func() { tTranscript = time.Since(t0).Milliseconds() }()
		if stdin.TranscriptPath != "" {
			transcData, _ = transcript.Parse(stdin.TranscriptPath)
		}
		return nil
	})

	g.Go(func() error {
		t0 := time.Now()
		defer func() { tNpm = time.Since(t0).Milliseconds() }()
		if stdin.Version != "" {
			home, _ := os.UserHomeDir()
			client := &npm.VersionClient{
				CacheDir: filepath.Join(home, ".claude", "ccsl", "cache"),
			}
			versionInfo = client.Check(stdin.Version)
		}
		return nil
	})

	if !cfg.Disabled["env"] && !cfg.Disabled["env.memory_usage"] {
		g.Go(func() error {
			t0 := time.Now()
			defer func() { tMemory = time.Since(t0).Milliseconds() }()
			memoryPct = fetchMemoryPct()
			return nil
		})
	}

	_ = g.Wait()
	fetchDone := time.Now()

	// Merge timing into fetchDetail after all goroutines complete (no concurrent writes).
	if debugOn {
		fetchDetail["git"] = tGit
		fetchDetail["transcript"] = tTranscript
		fetchDetail["npm_version"] = tNpm
		fetchDetail["memory"] = tMemory
		for k, v := range gitTimingLocal {
			fetchDetail[k] = v
		}
	}

	// 4. Build render context.
	styleMode := segment.StyleMode(cfg.Style)
	registry := segment.DefaultRegistry()
	sep := cfg.Separator
	if !cfg.SeparatorSet && sep == "" {
		sep = style.DefaultSeparator(styleMode)
	}

	// 5. Detect terminal width, reserving space for Claude Code's right-side
	// notification area (token count, version info, update status, etc.).
	// The notification area uses flex layout with no fixed width; 42 columns
	// covers the worst case (verbose tokens + version + update status) plus
	// paddingX=2 and gap=1 from PromptInputFooter.
	const rightReserved = 42
	termWidthRaw := render.TerminalWidth(120)
	termWidth := termWidthRaw - rightReserved
	if termWidth < 20 {
		termWidth = 20
	}

	ctx := &segment.RenderContext{
		Stdin:      stdin,
		Transcript: transcData,
		Git:        gitData,
		NpmVersion: versionInfo,
		MemoryPct:  memoryPct,
		Style:      styleMode,
		Disabled:   cfg.Disabled,
	}

	// 6. Build elements using shared pipeline.
	renderStart := time.Now()
	allElements := pipeline.BuildElements(cfg.Lines, registry, ctx, cfg.ThemeColors, styleMode)

	// Use RenderAllLines for cross-line alignment.
	outputLines, alignDiagram := render.RenderAllLines(allElements, sep, termWidth)
	renderDone := time.Now()

	// 7. Debug log (no-op if debug mode not enabled).
	if debugOn {
		var transcSummary *debug.TranscriptSummary
		if transcData != nil {
			var lastEvt string
			if !transcData.LastResponseTime.IsZero() {
				lastEvt = transcData.LastResponseTime.UTC().Format(time.RFC3339)
			}
			transcSummary = &debug.TranscriptSummary{
				LastResponseTime:    lastEvt,
				ActiveTools:         len(transcData.ActiveTools),
				CompletedTools:      transcData.CompletedTools,
				ToolStats:           transcData.ToolStats,
				ActiveAgents:        len(transcData.ActiveAgents),
				AgentStats:          transcData.AgentStats,
				Todos:               len(transcData.Todos),
				TodoCompleted:       transcData.TodoCompleted,
				Skills:              transcData.Skills,
				InputTokens:         transcData.UsageStats.InputTokens,
				OutputTokens:        transcData.UsageStats.OutputTokens,
				CacheCreationTokens: transcData.UsageStats.CacheCreationInputTokens,
				CacheReadTokens:     transcData.UsageStats.CacheReadInputTokens,
			}
		}
		debug.Log(&debug.Entry{
			TermWidthRaw:  termWidthRaw,
			TermWidthUsed: termWidth,
			RightReserved: rightReserved,
			ConfigStyle:   cfg.Style,
			ConfigLines:   len(cfg.Lines),
			ColorLevel:    config.DetectColorLevel(),
			Disabled:      cfg.Disabled,
			Stdin:         rawStdin,
			Git:           gitData,
			Transcript:    transcSummary,
			NpmVersion:    versionInfo,
			MemoryPct:     memoryPct,
			SmartAlign:    alignDiagram,
			OutputLines:   outputLines,
			Timing: &debug.Timing{
				StdinMs:     stdinDone.Sub(startTime).Milliseconds(),
				ConfigMs:    configDone.Sub(stdinDone).Milliseconds(),
				FetchMs:     fetchDone.Sub(configDone).Milliseconds(),
				FetchDetail: fetchDetail,
				RenderMs:    renderDone.Sub(renderStart).Milliseconds(),
			},
			DurationMs: time.Since(startTime).Milliseconds(),
		})
	}

	// 8. Output.
	if len(outputLines) > 0 {
		fmt.Print(strings.Join(outputLines, "\n"))
	}
	return nil
}

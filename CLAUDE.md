# CCSL — Claude Code StatusLine

## Build & Test

- `make build` — build binary
- `make check` — fmt + vet + lint + test (race-enabled, pre-commit gate)
- `make install` / `make i` — `go install` to GOPATH/bin
- `go test ./...` — run all unit tests
- `go test -tags integration -v .` — run end-to-end integration test
- `golangci-lint run` — lint (errcheck enabled). Requires v2 (`.golangci.yml` schema). Install: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`

## Architecture

- Single-process pipe model: CC pipes JSON via stdin → ccsl renders ANSI output to stdout; interactive TTY → config TUI
- `cmd/statusline.go` — main pipeline: stdin → concurrent fetch → render → stdout
- `cmd/config.go` — TUI config entry point (bubbletea)
- `internal/segment/` — ~50 segments (16 groups + ~35 leaf segments), each implements `Segment` interface
- `internal/render/` — ANSI primitives + line renderer with width budget + SmartAlign (see [docs/smartalign.md](docs/smartalign.md))
- `internal/transcript/` — JSONL parser with event normalization
- `internal/style/` — Nerd Font icon mapping + ANSI color styles
- `internal/config/` — config file loading and preset management
- `internal/tui/` — bubbletea TUI components for config editor
- `internal/api/` — Anthropic API client for usage/billing data
- `internal/git/` — git status helpers (branch, diff stats)
- `internal/input/` — stdin reader and JSON message parsing
- `internal/oauth/` — OAuth token extraction for API auth
- `internal/npm/` — npm (Claude Code) version checking
- `internal/pipeline/` — shared render pipeline (used by statusline and TUI preview)
- `themes/` — 8 embedded color themes (with block variants), go:embed in `themes/themes.go`
- `presets/` — 4 embedded preset configs (minimal/standard/full/dev)

## Conventions

- JSON library: `github.com/goccy/go-json` (use `gojson` import alias, use `gojson.RawMessage` not `encoding/json.RawMessage`)
- Import cycle between segment ↔ style resolved via `IconResolver` function pointer in `segment.go`, registered by `style.init()`
- Nerd Font codepoints: use Go `\U000XXXXX` syntax, NOT JS-style surrogate pairs (`\uDBxx\uDExx`)
- Every `Segment.Render()` must nil-check its data source and return `nil, nil` for missing data (graceful degradation)
- go:embed for presets lives in `presets/presets.go` (path relative to that file)
- Golden tests in `internal/render/testdata/` — regenerate with `go test ./internal/render/ -run TestGolden -update`
- Variable-width segments: `SegmentOutput.IsVariable=true` + `MinWidth` enables truncation by the width budget algorithm
- Segment disable key format: `"group.child"` (e.g., `"git.branch"`) in config's disabled map
- Pre-GA project: no backward compatibility concerns — refactor aggressively, no migration/compat shims needed
- Debug logging: all non-sensitive intermediate data (fetched values, computed state) should be included in the debug log entry (`internal/debug/`) for diagnostics
- **New group checklist**: adding a segment group requires updates in 4 places — `registry_default.go` (register), `internal/style/style.go` (icon), all 15 `themes/*.json` (color), and relevant `presets/*.json` (layout)
- Local-only docs (gitignored): `docs/superpowers/`, `docs/plans/`, `docs/brainstorms/` — superpowers skill specs/plans live here but are NOT committed
- Release tags: annotated, format `vX.Y.Z: <one-liner summary>` (e.g. `v0.3.0: channel-aware upgrade notice`)

package git

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

const gitTimeout = 1 * time.Second

// GitData holds parsed git repository status information.
type GitData struct {
	Branch     string `json:"branch"`
	Ahead      int    `json:"ahead"`
	Behind     int    `json:"behind"`
	HasRemote  bool   `json:"has_remote"`
	Modified   int    `json:"modified"`
	Added      int    `json:"added"`
	Deleted    int    `json:"deleted"`
	Untracked  int    `json:"untracked"`
	Insertions int    `json:"insertions"`
	Deletions  int    `json:"deletions"`
	IsClean    bool   `json:"is_clean"`
}

// runGit executes a git command in the given directory with a timeout.
func runGit(ctx context.Context, cwd string, args ...string) (string, error) {
	fullArgs := append([]string{"--no-optional-locks"}, args...)
	cmd := exec.CommandContext(ctx, "git", fullArgs...)
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\n\r"), nil
}

// GetData collects git repository data for the given working directory.
// Returns nil, nil if the directory is not inside a git repository.
func GetData(cwd string, timing map[string]int64) (*GitData, error) {
	data := &GitData{}
	g, ctx := errgroup.WithContext(context.Background())

	// Per-goroutine timing captured in local vars to avoid concurrent map writes.
	var tStatus, tDiff int64

	// 1. git status --porcelain=v2 -b → branch, ahead/behind, file changes.
	// This also serves as the git-repo check: if it fails, we're not in a repo.
	var notGitRepo bool
	g.Go(func() error {
		t0 := time.Now()
		defer func() { tStatus = time.Since(t0).Milliseconds() }()
		ctx, cancel := context.WithTimeout(ctx, gitTimeout)
		defer cancel()
		out, err := runGit(ctx, cwd, "status", "--porcelain=v2", "-b")
		if err != nil {
			notGitRepo = true
			return nil
		}
		parseStatusV2(out, data)

		// Detached HEAD: branch.head is "(detached)", resolve to short hash.
		if data.Branch == "(detached)" {
			short, err := runGit(ctx, cwd, "rev-parse", "--short", "HEAD")
			if err != nil {
				data.Branch = "(detached)"
			} else {
				data.Branch = short
			}
		}
		return nil
	})

	// 2. Diff numstat → Insertions, Deletions.
	g.Go(func() error {
		t0 := time.Now()
		defer func() { tDiff = time.Since(t0).Milliseconds() }()
		ctx, cancel := context.WithTimeout(ctx, gitTimeout)
		defer cancel()
		out, err := runGit(ctx, cwd, "diff", "--numstat", "HEAD")
		if err != nil {
			// Fallback: cached diff (works on unborn HEAD).
			out, err = runGit(ctx, cwd, "diff", "--numstat", "--cached")
			if err != nil {
				return nil
			}
		}
		ins, del := parseNumstat(out)
		data.Insertions = ins
		data.Deletions = del
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("git data: %w", err)
	}

	if notGitRepo {
		return nil, nil
	}

	if timing != nil {
		timing["git.status"] = tStatus
		timing["git.diff"] = tDiff
	}

	return data, nil
}

// parseStatusV2 parses the output of `git status --porcelain=v2 -b` into GitData.
func parseStatusV2(out string, data *GitData) {
	if out == "" {
		data.Branch = "(init)"
		data.IsClean = true
		return
	}

	hasChanges := false

	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "# branch.head "):
			data.Branch = strings.TrimPrefix(line, "# branch.head ")

		case strings.HasPrefix(line, "# branch.ab "):
			// Format: "# branch.ab +<ahead> -<behind>"
			data.HasRemote = true
			ab := strings.TrimPrefix(line, "# branch.ab ")
			parts := strings.Fields(ab)
			if len(parts) == 2 {
				// parts[0] = "+N", parts[1] = "-N"
				data.Ahead, _ = strconv.Atoi(parts[0][1:])
				data.Behind, _ = strconv.Atoi(parts[1][1:])
			}

		case strings.HasPrefix(line, "# "):
			// Other header lines (branch.oid, branch.upstream) — skip.
			continue

		case line[0] == '?':
			// Untracked file.
			data.Untracked++
			hasChanges = true

		case line[0] == '1' || line[0] == '2' || line[0] == 'u':
			// Ordinary change (1), rename/copy (2), or unmerged (u).
			// XY is at position 2-3 (after "1 " / "2 " / "u ").
			if len(line) < 4 {
				continue
			}
			xy := line[2:4]
			switch {
			case xy[0] == 'A' || xy[1] == 'A':
				data.Added++
			case xy[0] == 'D' || xy[1] == 'D':
				data.Deleted++
			case xy[0] == 'M' || xy[1] == 'M' ||
				xy[0] == 'R' || xy[1] == 'R' ||
				xy[0] == 'C' || xy[1] == 'C' ||
				xy[0] == 'U' || xy[1] == 'U':
				data.Modified++
			}
			hasChanges = true
		}
	}

	if data.Branch == "" {
		data.Branch = "(init)"
	}
	data.IsClean = !hasChanges
}

// parseNumstat sums insertions and deletions from `git diff --numstat` output.
func parseNumstat(out string) (ins, del int) {
	if out == "" {
		return 0, 0
	}
	for _, line := range strings.Split(out, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		// Binary files show "-" for counts.
		if parts[0] == "-" || parts[1] == "-" {
			continue
		}
		i, _ := strconv.Atoi(parts[0])
		d, _ := strconv.Atoi(parts[1])
		ins += i
		del += d
	}
	return ins, del
}

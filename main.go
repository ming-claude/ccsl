package main

import (
	"fmt"
	"os"

	"github.com/ming-claude/ccsl/cmd"
)

var version = "dev"

func main() {
	// --version / -v: print version and exit.
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println(version)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "config" {
		runConfig()
		return
	}

	// No subcommand: auto-detect mode.
	// TTY stdin (user ran `ccsl` directly) → config TUI.
	// Piped stdin (CC invocation) → statusline pipeline.
	if isStdinTTY() {
		runConfig()
		return
	}

	if err := cmd.RunStatusline(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runConfig() {
	if err := cmd.RunConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func isStdinTTY() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

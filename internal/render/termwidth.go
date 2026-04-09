package render

import (
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

// TerminalWidth detects the terminal column count.
// In pipe mode (stdin/stdout redirected), it falls back through:
// stderr → /dev/tty → COLUMNS env var → defaultWidth.
func TerminalWidth(defaultWidth int) int {
	// Try stderr first — often still connected to TTY in pipe mode.
	if w := ioctlWidth(os.Stderr.Fd()); w > 0 {
		return w
	}
	// Try stdout (works when not piped).
	if w := ioctlWidth(os.Stdout.Fd()); w > 0 {
		return w
	}
	// Try /dev/tty directly.
	if f, err := os.Open("/dev/tty"); err == nil {
		w := ioctlWidth(f.Fd())
		_ = f.Close()
		if w > 0 {
			return w
		}
	}
	// COLUMNS env var.
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(cols); err == nil && n > 0 {
			return n
		}
	}
	return defaultWidth
}

func ioctlWidth(fd uintptr) int {
	ws, err := unix.IoctlGetWinsize(int(fd), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 {
		return 0
	}
	return int(ws.Col)
}

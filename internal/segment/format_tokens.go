package segment

import (
	"fmt"
	"strings"
	"time"
)

// formatTokens formats a token count with adaptive units, always
// maintaining 3 significant digits: 8.44M, 12.3k, 1k, 500.
func formatTokens(n int) string {
	if n >= 1_000_000 {
		return format3Sig(float64(n)/1_000_000, "M")
	}
	if n >= 1000 {
		return format3Sig(float64(n)/1000, "k")
	}
	return fmt.Sprintf("%d", n)
}

// format3Sig formats a float with up to 3 significant digits plus a unit
// suffix, trimming trailing zeros after the decimal point.
func format3Sig(v float64, unit string) string {
	var s string
	switch {
	case v >= 100:
		return fmt.Sprintf("%.0f%s", v, unit)
	case v >= 10:
		s = fmt.Sprintf("%.1f", v)
	default:
		s = fmt.Sprintf("%.2f", v)
	}
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s + unit
}

// formatCountdown formats a duration compactly as "XdXh", "XhXm", or "Xm".
func formatCountdown(d time.Duration) string {
	totalMinutes := int(d.Minutes())
	days := totalMinutes / (24 * 60)
	hours := (totalMinutes / 60) % 24
	minutes := totalMinutes % 60
	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

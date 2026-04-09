//go:build darwin

package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// fetchMemoryPct runs vm_stat and returns the memory usage percentage string.
func fetchMemoryPct() string {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	out, err := exec.CommandContext(ctx, "vm_stat").Output()
	if err != nil {
		return ""
	}

	stats := parseVMStat(string(out))
	free := stats["Pages free"]
	active := stats["Pages active"]
	inactive := stats["Pages inactive"]
	speculative := stats["Pages speculative"]
	wired := stats["Pages wired down"]

	total := free + active + inactive + speculative + wired
	if total == 0 {
		return ""
	}

	used := total - free
	pct := float64(used) / float64(total) * 100
	return fmt.Sprintf("%.0f%%", pct)
}

// parseVMStat parses vm_stat output into a map of name → page count.
func parseVMStat(output string) map[string]int64 {
	stats := make(map[string]int64)
	for _, line := range strings.Split(output, "\n") {
		idx := strings.LastIndex(line, ":")
		if idx < 0 {
			continue
		}
		name := strings.TrimSpace(line[:idx])
		valStr := strings.TrimSpace(line[idx+1:])
		valStr = strings.TrimSuffix(valStr, ".")
		val, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			continue
		}
		stats[name] = val
	}
	return stats
}

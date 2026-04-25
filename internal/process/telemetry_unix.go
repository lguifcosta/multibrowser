//go:build !windows

package process

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// sampleTelemetry reads CPU% and RAM for a given PID.
// prevProcJiff/prevTotalJiff are the previous readings (pass 0 on first call).
// Returns cpuPercent (0 on first call), ramMB, and updated jiffies for next call.
func sampleTelemetry(pid int, prevProcJiff, prevTotalJiff uint64) (cpuPct float64, ramMB float64, curProcJiff, curTotalJiff uint64) {
	statData, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, 0, prevProcJiff, prevTotalJiff
	}

	// Skip the comm field (may contain spaces) by finding the last ')'
	s := string(statData)
	idx := strings.LastIndex(s, ")")
	if idx < 0 {
		return 0, 0, prevProcJiff, prevTotalJiff
	}
	parts := strings.Fields(strings.TrimSpace(s[idx+1:]))
	// After ')': state(0) ppid(1) pgrp(2) session(3) tty(4) tpgid(5) flags(6)
	//             minflt(7) cminflt(8) majflt(9) cmajflt(10) utime(11) stime(12)
	if len(parts) < 13 {
		return 0, 0, prevProcJiff, prevTotalJiff
	}
	utime, _ := strconv.ParseUint(parts[11], 10, 64)
	stime, _ := strconv.ParseUint(parts[12], 10, 64)
	curProcJiff = utime + stime

	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, curProcJiff, prevTotalJiff
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		for _, field := range strings.Fields(scanner.Text())[1:] {
			v, _ := strconv.ParseUint(field, 10, 64)
			curTotalJiff += v
		}
	}

	if prevProcJiff > 0 && curTotalJiff > prevTotalJiff && curProcJiff >= prevProcJiff {
		procDelta := float64(curProcJiff - prevProcJiff)
		totalDelta := float64(curTotalJiff - prevTotalJiff)
		cpuPct = (procDelta / totalDelta) * 100.0
	}

	statusData, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return cpuPct, 0, curProcJiff, curTotalJiff
	}
	for _, line := range strings.Split(string(statusData), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, _ := strconv.ParseFloat(fields[1], 64)
				ramMB = kb / 1024.0
			}
			break
		}
	}

	return cpuPct, ramMB, curProcJiff, curTotalJiff
}

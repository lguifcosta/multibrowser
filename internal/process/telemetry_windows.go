//go:build windows

package process

// sampleTelemetry is a stub on Windows — returns zeroes.
func sampleTelemetry(pid int, prevProcJiff, prevTotalJiff uint64) (cpuPct float64, ramMB float64, curProcJiff, curTotalJiff uint64) {
	return 0, 0, 0, 0
}

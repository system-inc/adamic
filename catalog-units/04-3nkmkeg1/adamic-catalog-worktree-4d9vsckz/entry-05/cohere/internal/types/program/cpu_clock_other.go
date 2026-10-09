//go:build !(darwin || linux)

package program

import "time"

// threadCPU has no clock to read here, so --timing says its CPU columns are unavailable rather than
// printing zeros. See cpu_clock_unix.go.
func threadCPU() (time.Duration, bool) {
	return 0, false
}

// processCPU has no clock to read here either.
func processCPU() (time.Duration, bool) {
	return 0, false
}

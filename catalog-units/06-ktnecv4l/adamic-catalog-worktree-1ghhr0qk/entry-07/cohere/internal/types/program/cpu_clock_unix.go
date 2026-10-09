//go:build darwin || linux

package program

import (
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// threadCPU is the CPU the calling thread has used, user and system, by the kernel's own count. False
// when the clock cannot be read.
//
// Measured on darwin/arm64 (2026-10-04): about 100ns a read, in steps of 41ns, and a 50ms sleep costs
// the sleeping thread 28µs of it. Waiting costs nothing, which is the whole reason --timing reads this
// rather than the wall clock (#8qyzmxw).
func threadCPU() (time.Duration, bool) {
	var spec unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_THREAD_CPUTIME_ID, &spec); err != nil {
		return 0, false
	}
	return time.Duration(spec.Nano()), true
}

// processCPU is the CPU the whole process has used, user and system. False when it cannot be read.
func processCPU() (time.Duration, bool) {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return 0, false
	}
	return time.Duration(usage.Utime.Nano() + usage.Stime.Nano()), true
}

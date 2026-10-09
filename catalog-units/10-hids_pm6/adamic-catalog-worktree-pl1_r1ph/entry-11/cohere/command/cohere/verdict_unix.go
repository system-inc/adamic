//go:build !windows

package main

import (
	"os"
	"strconv"
	"syscall"
)

func takeVerdictFile() *os.File {
	value := os.Getenv(VerdictVariable)
	if value == "" {
		return nil
	}
	os.Unsetenv(VerdictVariable)
	descriptor, err := strconv.Atoi(value)
	if err != nil || descriptor < 3 {
		return nil
	}
	// Only a pipe is the dispatcher's. The variable can reach an engine from outside, a user's environment
	// or a script, where the same number may be a file the script opened (`3>log`), and a byte written
	// there would land in someone else's output. Such a descriptor is left exactly as it was found.
	var status syscall.Stat_t
	if err := syscall.Fstat(descriptor, &status); err != nil || status.Mode&syscall.S_IFMT != syscall.S_IFIFO {
		return nil
	}
	syscall.CloseOnExec(descriptor)
	return os.NewFile(uintptr(descriptor), "verdict")
}

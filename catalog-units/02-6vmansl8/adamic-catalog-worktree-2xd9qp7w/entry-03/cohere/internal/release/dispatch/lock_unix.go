//go:build unix

package dispatch

import (
	"fmt"
	"os"
	"syscall"
)

// lockExclusive takes an exclusive lock on path, waiting for any holder, and returns its release.
//
// flock is held by the open file, so the kernel drops it when the holder exits, however it exits. A
// lock that a crashed build could leave behind would need staleness rules, and any staleness rule is a
// guess about how long a cold Swift build takes.
func lockExclusive(path string) (func(), error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("opening the lock %s: %w", path, err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return nil, fmt.Errorf("taking the lock %s: %w", path, err)
	}
	return func() {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		file.Close()
	}, nil
}

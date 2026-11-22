package childguard

import (
	"syscall"
	"unsafe"
)

// waitidPid is waitid's `P_PID`: the id names one process.
const waitidPid = 1

// awaitExit blocks until the leader has exited and reports true, leaving it unreaped (`WNOWAIT`), so it still holds its
// pid and its group's id. It reports false when waitid fails, and the caller then waits by reaping.
func awaitExit(leader int) bool {
	return waitid(leader, syscall.WEXITED|syscall.WNOWAIT) == nil
}

// waitid is waitid(2) on one process, retried through `EINTR`. The syscall package doesn't wrap it, so this calls it as
// the standard library's internal/syscall/unix does, with siginfo_t's 128 bytes to fill.
func waitid(pid, options int) error {
	var info [16]uint64
	for {
		_, _, errno := syscall.Syscall6(syscall.SYS_WAITID, waitidPid, uintptr(pid), uintptr(unsafe.Pointer(&info)), uintptr(options), 0, 0)
		if errno != syscall.EINTR {
			if errno != 0 {
				return errno
			}
			return nil
		}
	}
}

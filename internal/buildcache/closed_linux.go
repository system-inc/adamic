//go:build linux && !mips && !mipsle && !mips64 && !mips64le

package buildcache

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// The kernel's numbers closedProof asks with, which package syscall doesn't export. faccessat2's call number is 439 on
// every architecture this file builds for; mips numbers its calls apart, so it builds closed_other.go instead. W_OK is
// writeAccess (system.go).
const (
	statfsReadOnly      = 0x1   // ST_RDONLY
	faccessat2Call      = 439   // SYS_faccessat2, Linux 5.8
	effectiveIdentities = 0x200 // AT_EACCESS
	currentDirectory    = -100  // AT_FDCWD
)

// closedProof is how root is proven closed to this process, or why it isn't: closedProofWith over this machine's own
// reads.
func closedProof(root string) (string, error) {
	return closedProofWith(root, closureReads{
		readOnlyMount: func(path string) bool {
			var mount syscall.Statfs_t
			return syscall.Statfs(path, &mount) == nil && uint64(mount.Flags)&statfsReadOnly != 0
		},
		mountInfo: func() ([]byte, error) { return os.ReadFile("/proc/self/mountinfo") },
		status:    func() ([]byte, error) { return os.ReadFile("/proc/self/status") },
		owner:     ownerOf,
		writable:  writableByThisProcess,
	})
}

// ownerOf is the uid owning path, the path itself and never what a link names (lstat).
func ownerOf(path string) (int, error) {
	var stat syscall.Stat_t
	if err := syscall.Lstat(path, &stat); err != nil {
		return 0, err
	}
	return int(stat.Uid), nil
}

// writableByThisProcess asks the kernel whether this process's effective uid and groups may write path: faccessat2 with
// AT_EACCESS, called directly rather than through syscall.Faccessat, which tries faccessat2 too but, where the kernel or
// a seccomp filter refuses it, judges by mode bits instead, which an ACL granting a write gets past. The kernel's
// answer weighs ACLs and a read-only mount (EROFS). EPERM is faccessat2 refused by a seccomp filter as often as an
// immutable file, so it proves nothing, as ENOSYS doesn't.
func writableByThisProcess(path string) (bool, error) {
	switch err := faccessat2(path, writeAccess, effectiveIdentities); {
	case err == nil:
		return true, nil
	case errors.Is(err, syscall.EACCES), errors.Is(err, syscall.EROFS):
		return false, nil
	case errors.Is(err, syscall.ENOSYS), errors.Is(err, syscall.EPERM):
		return false, fmt.Errorf("faccessat2 answers %v, and mode bits alone would miss an ACL that grants a write", err)
	default:
		return false, err
	}
}

// faccessat2 asks the kernel whether this process may access path with mode, relative to the working directory.
func faccessat2(path string, mode, flags int) error {
	name, err := syscall.BytePtrFromString(path)
	if err != nil {
		return err
	}
	// A variable, since the constant -100 can't convert to uintptr; the kernel reads the word back as an int.
	directory := currentDirectory
	if _, _, errno := syscall.Syscall6(faccessat2Call, uintptr(directory), uintptr(unsafe.Pointer(name)), uintptr(mode), uintptr(flags), 0, 0); errno != 0 {
		return errno
	}
	return nil
}

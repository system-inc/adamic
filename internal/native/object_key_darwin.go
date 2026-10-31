package native

import (
	"io/fs"
	"syscall"
)

// fileChange is the inode of the file info describes and when the inode last changed, its ctime, in nanoseconds
// (objectReadOf).
func fileChange(info fs.FileInfo) (uint64, int64, bool) {
	system, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, false
	}
	return system.Ino, system.Ctimespec.Nano(), true
}

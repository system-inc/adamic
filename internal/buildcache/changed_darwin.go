package buildcache

import (
	"io/fs"
	"syscall"
	"time"
)

// changeTime is when a file's inode last changed, its ctime (cleanObject).
func changeTime(info fs.FileInfo) (time.Time, bool) {
	system, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return time.Time{}, false
	}
	return time.Unix(system.Ctimespec.Unix()), true
}

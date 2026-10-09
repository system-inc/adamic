package program

import (
	"os"
	"syscall"
)

// statIdentity is what stat says identifies a file's current bytes, following symlinks as a read does.
func statIdentity(path string) (fileIdentity, bool) {
	var status syscall.Stat_t
	if syscall.Stat(path, &status) != nil {
		return fileIdentity{}, false
	}
	return fileIdentity{
		size:                status.Size,
		modifiedNanoseconds: status.Mtimespec.Nano(),
		changedNanoseconds:  status.Ctimespec.Nano(),
		inode:               status.Ino,
		device:              uint64(status.Dev),
	}, true
}

// changeTimeAndInode reads a stat's change time and inode, which os.FileInfo does not carry itself.
func changeTimeAndInode(path string, information os.FileInfo) (changedNanoseconds int64, inode uint64) {
	status, ok := information.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0
	}
	return status.Ctimespec.Nano(), status.Ino
}

// identityFromInfo is statIdentity read from a stat already taken, so a caller that statted a file for another
// reason need not stat it again.
func identityFromInfo(information os.FileInfo) (fileIdentity, bool) {
	status, ok := information.Sys().(*syscall.Stat_t)
	if !ok {
		return fileIdentity{}, false
	}
	return fileIdentity{
		size:                status.Size,
		modifiedNanoseconds: status.Mtimespec.Nano(),
		changedNanoseconds:  status.Ctimespec.Nano(),
		inode:               status.Ino,
		device:              uint64(status.Dev),
	}, true
}

//go:build !darwin && !linux && !windows

package program

import "os"

// statIdentity reports nothing here, so the content pack serves nothing and every file is read from disk.
// A change time is what makes a pack entry safe to serve, and these platforms do not give one through the
// same call.
func statIdentity(path string) (fileIdentity, bool) {
	return fileIdentity{}, false
}

// changeTimeAndInode reports neither here, so a run cache input is signed by its size and modification time
// alone, as it was before either was added.
func changeTimeAndInode(path string, information os.FileInfo) (changedNanoseconds int64, inode uint64) {
	return 0, 0
}

// identityFromInfo reports nothing here, as statIdentity does.
func identityFromInfo(information os.FileInfo) (fileIdentity, bool) {
	return fileIdentity{}, false
}

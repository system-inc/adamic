package program

import (
	"os"
	"syscall"
	"unsafe"
)

// statIdentity reports nothing on Windows, so the content pack serves nothing and every file is read from
// disk. A pack entry is safe to serve only on a change time read in the same call as the rest, and Windows
// gives one only through an open handle, which would cost more per file than the read it saves.
func statIdentity(path string) (fileIdentity, bool) {
	return fileIdentity{}, false
}

// fileBasicInfo is Windows' FILE_BASIC_INFO, laid out field for field. Its times are FILETIMEs: hundreds
// of nanoseconds since 1601.
type fileBasicInfo struct {
	creationTime   int64
	lastAccessTime int64
	lastWriteTime  int64
	changeTime     int64
	fileAttributes uint32
}

// fileBasicInfoClass is FileBasicInfo in FILE_INFO_BY_HANDLE_CLASS.
const fileBasicInfoClass = 0

// fileReadAttributes is FILE_READ_ATTRIBUTES: opening for it reads no data, so it is the cheapest open
// Windows allows and one an antivirus filter has no content to scan for.
const fileReadAttributes = 0x80

// unixEpochInFiletime is 1970-01-01 as a FILETIME.
const unixEpochInFiletime = 116444736000000000

var getFileInformationByHandleEx = syscall.NewLazyDLL("kernel32.dll").NewProc("GetFileInformationByHandleEx")

// changeTimeAndInode reads a file's change time and file index, Windows' counterparts of a stat's ctime and
// inode, which its os.FileInfo does not carry. They catch what the modification time cannot, and on Windows
// that case is ordinary rather than exotic: Explorer, Copy-Item, xcopy and robocopy all copy a file's
// modification time along with its bytes.
//
// Both come through an open handle, opened for attributes only and sharing everything, so it never stands
// in another process's way. Either answers zero when Windows does not give it, and an input whose open
// failed is then signed by its size and modification time alone, as on a platform with neither.
func changeTimeAndInode(path string, information os.FileInfo) (changedNanoseconds int64, inode uint64) {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0
	}
	handle, err := syscall.CreateFile(name, fileReadAttributes,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE, nil,
		// Backup semantics is what lets CreateFile open a directory, which the run cache records too.
		syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return 0, 0
	}
	defer syscall.CloseHandle(handle)

	var basic fileBasicInfo
	if succeeded, _, _ := getFileInformationByHandleEx.Call(uintptr(handle), fileBasicInfoClass,
		uintptr(unsafe.Pointer(&basic)), unsafe.Sizeof(basic)); succeeded != 0 {
		changedNanoseconds = (basic.changeTime - unixEpochInFiletime) * 100
	}
	var identity syscall.ByHandleFileInformation
	if syscall.GetFileInformationByHandle(handle, &identity) == nil {
		inode = uint64(identity.FileIndexHigh)<<32 | uint64(identity.FileIndexLow)
	}
	return changedNanoseconds, inode
}

// identityFromInfo reports nothing here, as statIdentity does.
func identityFromInfo(information os.FileInfo) (fileIdentity, bool) {
	return fileIdentity{}, false
}

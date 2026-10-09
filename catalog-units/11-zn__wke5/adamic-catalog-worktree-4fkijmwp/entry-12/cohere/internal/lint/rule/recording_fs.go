package rule

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// FileRead is one path a reader asked the file system about, and whether it was there.
//
// Absence is recorded as carefully as presence: a design system that looked for an entry point and did
// not find it depends on that file staying absent, since creating it changes the design system.
//
// A present path keeps the size and modification time it had when it was first asked about, so a
// cache hashing it afterwards can re-stat and decline to record a file that moved between the read and
// the hash, the race the run cache closes the same way for its own inputs.
type FileRead struct {
	Path                string
	Present             bool
	Size                int64
	ModifiedNanoseconds int64
}

// RecordingFS is a file system that notes every path it is asked about and refuses to write.
//
// It is how a reader's inputs are observed rather than listed, the discipline the run cache's input
// recorder holds the compiler to, applied to one reader: whatever the design system asked about is in
// its read set, so a cache keyed on that set cannot miss a file it forgot (#pyhm2t2). Writes panic,
// because a reader that writes is not a reader, and a cached verdict cannot replay a side effect.
type RecordingFS struct {
	underlying vfs.FS
	mutex      sync.Mutex
	reads      map[string]FileRead
}

var _ vfs.FS = (*RecordingFS)(nil)

// NewRecordingFS wraps underlying.
func NewRecordingFS(underlying vfs.FS) *RecordingFS {
	return &RecordingFS{underlying: underlying, reads: map[string]FileRead{}}
}

// Reads is every path asked about, sorted. A path seen both present and absent keeps present.
func (recording *RecordingFS) Reads() []FileRead {
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	reads := make([]FileRead, 0, len(recording.reads))
	for _, read := range recording.reads {
		reads = append(reads, read)
	}
	sort.Slice(reads, func(first, second int) bool { return reads[first].Path < reads[second].Path })
	return reads
}

func (recording *RecordingFS) note(path string, present bool) {
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	if earlier, seen := recording.reads[path]; seen && (earlier.Present || !present) {
		return
	}
	read := FileRead{Path: path, Present: present}
	if present {
		if info := recording.underlying.Stat(tspath.RootedPath(path)); info != nil {
			read.Size = info.Size()
			read.ModifiedNanoseconds = info.ModTime().UnixNano()
		}
	}
	recording.reads[path] = read
}

func (recording *RecordingFS) refuse(method string, path string) {
	panic(fmt.Sprintf("a recording file system refuses %s(%s): its reader must not write", method, path))
}

func (recording *RecordingFS) CaseSensitivity() tspath.CaseSensitivity {
	return recording.underlying.CaseSensitivity()
}

func (recording *RecordingFS) FileExists(path tspath.RootedFilePath) bool {
	exists := recording.underlying.FileExists(path)
	recording.note(path.AsString(), exists)
	return exists
}

func (recording *RecordingFS) ReadFile(path tspath.RootedFilePath) (string, bool) {
	contents, ok := recording.underlying.ReadFile(path)
	recording.note(path.AsString(), ok)
	return contents, ok
}

func (recording *RecordingFS) WriteFile(path tspath.RootedFilePath, data string) error {
	recording.refuse("WriteFile", path.AsString())
	return nil
}

func (recording *RecordingFS) AppendFile(path tspath.RootedFilePath, data string) error {
	recording.refuse("AppendFile", path.AsString())
	return nil
}

func (recording *RecordingFS) Remove(path tspath.RootedPath) error {
	recording.refuse("Remove", path.AsString())
	return nil
}

func (recording *RecordingFS) Chtimes(path tspath.RootedPath, accessTime time.Time, modificationTime time.Time) error {
	recording.refuse("Chtimes", path.AsString())
	return nil
}

func (recording *RecordingFS) DirectoryExists(path tspath.RootedDirectoryPath) bool {
	exists := recording.underlying.DirectoryExists(path)
	recording.note(path.AsString(), exists)
	return exists
}

func (recording *RecordingFS) GetAccessibleEntries(path tspath.RootedDirectoryPath) vfs.Entries {
	entries := recording.underlying.GetAccessibleEntries(path)
	recording.note(path.AsString(), len(entries.Files) > 0 || len(entries.Directories) > 0 || recording.underlying.DirectoryExists(path))
	return entries
}

func (recording *RecordingFS) Stat(path tspath.RootedPath) vfs.FileInfo {
	info := recording.underlying.Stat(path)
	recording.note(path.AsString(), info != nil)
	return info
}

func (recording *RecordingFS) Realpath(path tspath.RootedPath) tspath.RootedPath {
	recording.note(path.AsString(), recording.underlying.FileExists(tspath.RootedFilePath(path)) || recording.underlying.DirectoryExists(tspath.RootedDirectoryPath(path)))
	return tspath.RootedPath(recording.underlying.Realpath(path).AsString())
}

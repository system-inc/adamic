package lowered

import (
	"errors"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

var errSealed = errors.New("lowered: the sealed file system is read-only")

// sealed is the disk as a stored lowering's checker sees it: the program's sources, held in memory where they sit, and
// nothing else. Every other path the checker asks after (a package.json resolution looks for, a node_modules directory,
// a file an import might name) answers absent and is recorded, the trace of what the lowering's resolution read
// outside its sources. The checker never reads this machine's disk, so the lowering is the one its sources and that
// trace's absences give, wherever it's built.
type sealed struct {
	files       map[string]string
	directories map[string]bool
	mutex       sync.Mutex
	asked       map[string]bool
}

// sealedOver holds files, by absolute slash path, and every directory above each.
func sealedOver(files map[string]string) *sealed {
	disk := &sealed{files: files, directories: map[string]bool{}, asked: map[string]bool{}}
	for file := range files {
		for directory := path.Dir(file); !disk.directories[directory]; directory = path.Dir(directory) {
			disk.directories[directory] = true
			if directory == "/" {
				break
			}
		}
	}
	return disk
}

// ask records a path the sources don't hold.
func (disk *sealed) ask(name string) {
	disk.mutex.Lock()
	defer disk.mutex.Unlock()
	disk.asked[name] = true
}

// trace is every path asked after that the sources don't hold, sorted.
func (disk *sealed) trace() []string {
	disk.mutex.Lock()
	defer disk.mutex.Unlock()
	names := make([]string, 0, len(disk.asked))
	for name := range disk.asked {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// CaseSensitivity is the machine's, so names compare as a fresh lowering's do.
func (disk *sealed) CaseSensitivity() tspath.CaseSensitivity {
	return osvfs.FS().CaseSensitivity()
}

func (disk *sealed) FileExists(name tspath.RootedFilePath) bool {
	if _, held := disk.files[name.AsString()]; held {
		return true
	}
	disk.ask(name.AsString())
	return false
}

func (disk *sealed) ReadFile(name tspath.RootedFilePath) (string, bool) {
	if content, held := disk.files[name.AsString()]; held {
		return content, true
	}
	disk.ask(name.AsString())
	return "", false
}

func (disk *sealed) DirectoryExists(name tspath.RootedDirectoryPath) bool {
	if disk.directories[name.AsString()] {
		return true
	}
	disk.ask(name.AsString())
	return false
}

// GetAccessibleEntries lists what the sources hold under a directory. A listing is a read of the directory, so it's
// recorded too, as the directory with a trailing slash: on disk it may hold more than the sources.
func (disk *sealed) GetAccessibleEntries(name tspath.RootedDirectoryPath) vfs.Entries {
	directory := strings.TrimSuffix(name.AsString(), "/")
	disk.ask(directory + "/")
	var entries vfs.Entries
	for _, held := range []map[string]bool{disk.directories, filesOf(disk.files)} {
		for entry := range held {
			if entry != directory && path.Dir(entry) == directory {
				if _, file := disk.files[entry]; file {
					entries.Files = append(entries.Files, path.Base(entry))
				} else {
					entries.Directories = append(entries.Directories, path.Base(entry))
				}
			}
		}
	}
	sort.Strings(entries.Files)
	sort.Strings(entries.Directories)
	return entries
}

// filesOf is files' names as a set.
func filesOf(files map[string]string) map[string]bool {
	names := make(map[string]bool, len(files))
	for name := range files {
		names[name] = true
	}
	return names
}

func (disk *sealed) Stat(name tspath.RootedPath) vfs.FileInfo {
	if content, held := disk.files[name.AsString()]; held {
		return sealedInfo{name: path.Base(name.AsString()), size: int64(len(content))}
	}
	if disk.directories[name.AsString()] {
		return sealedInfo{name: path.Base(name.AsString()), directory: true}
	}
	disk.ask(name.AsString())
	return nil
}

// Realpath is the path itself: the sources hold no links.
func (disk *sealed) Realpath(name tspath.RootedPath) tspath.RootedPath {
	return name
}

func (disk *sealed) WriteFile(tspath.RootedFilePath, string) error  { return errSealed }
func (disk *sealed) AppendFile(tspath.RootedFilePath, string) error { return errSealed }
func (disk *sealed) Remove(tspath.RootedPath) error                 { return errSealed }
func (disk *sealed) Chtimes(tspath.RootedPath, time.Time, time.Time) error {
	return errSealed
}

// sealedInfo is a held file's or directory's stat, the same on every machine.
type sealedInfo struct {
	name      string
	size      int64
	directory bool
}

func (info sealedInfo) Name() string { return info.name }
func (info sealedInfo) Size() int64  { return info.size }
func (info sealedInfo) Mode() fs.FileMode {
	if info.directory {
		return fs.ModeDir | 0o755
	}
	return 0o644
}
func (info sealedInfo) ModTime() time.Time { return time.Time{} }
func (info sealedInfo) IsDir() bool        { return info.directory }
func (info sealedInfo) Sys() any           { return nil }

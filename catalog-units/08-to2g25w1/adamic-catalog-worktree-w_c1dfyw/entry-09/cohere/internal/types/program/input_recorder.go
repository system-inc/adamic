package program

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// InputRecorder notes every path the compiler read or looked for on disk during a build.
//
// It exists for the run cache, and the reason it observes rather than enumerates is that every
// hand-written list of a build's inputs turned out to be missing something. The program's source files
// are the obvious inputs. Less obvious: the tsconfig's extends chain, every package.json module
// resolution consulted, and every path resolution probed and did NOT find, since creating one of those
// changes what an import resolves to. An in-place edit to a package.json beside the tsconfig moves no
// directory's mtime, so a cache that only watched program files and their directories would replay a
// stale verdict over a changed resolution. Recording at the filesystem boundary sees all of them
// because it sees whatever the compiler did, not what someone remembered it does.
//
// It wraps the real disk beneath the memoizing layer, so each path reaches it about once, and beneath
// the embedded lib overlay, so the bundled lib.*.d.ts files never appear here: they are part of the
// binary, and the binary is already in the run cache's key.
type InputRecorder struct {
	mutex sync.Mutex

	// existed maps each path the build touched to whether it was there. A path seen both ways keeps
	// true: something the run read is an input that must keep its signature.
	existed map[string]bool

	// written is every path the build wrote, which is an output rather than an input. The types phase
	// reads its incremental build info at the start and writes it at the end; recorded as an input it
	// would change on every run, and every following check would miss while the cache looked merely
	// cold. Dropping it is sound because the incremental path already guarantees identical findings
	// warm or cold, so its contents cannot change the verdict being cached.
	written map[string]bool

	// depended is every path whose answer depended on more than its existence: a file read, a directory
	// listed, a path statted. Every other present path was only asked whether it is there, and only that
	// can change its answer (see Inputs).
	depended map[string]bool
}

// NewInputRecorder returns an empty recorder.
func NewInputRecorder() *InputRecorder {
	return &InputRecorder{existed: map[string]bool{}, written: map[string]bool{}, depended: map[string]bool{}}
}

func (r *InputRecorder) note(path string, present bool, depended bool) {
	if r == nil || path == "" {
		return
	}
	path = withoutTrailingSlash(path)
	r.mutex.Lock()
	r.existed[path] = r.existed[path] || present
	r.depended[path] = r.depended[path] || depended
	r.mutex.Unlock()
}

// Inputs reports what the build depended on: paths that existed, and paths it looked for and did not
// find. Anything the build itself wrote is excluded. Sorted so two recordings of one tree compare equal.
//
// probed is the present paths the build only asked whether they exist, never read, listed or statted, a
// subset of present. Module resolution asks that of every directory from the project up to the root, on its
// way to a package.json or a node_modules, and the answer can only change by the path ceasing to be there.
// See RecordRunCache, which records such a directory for its existence alone.
//
// The same tree records the same inputs on every build (#zc57tqg). Which existence probes reach the disk
// depends on which of the parallel loaders asks first: on excalidraw, six builds of one tree each recorded the
// probe of a package's directory, such as node_modules/react-dom/, only sometimes, the times module resolution
// answered a second import of the package from its own cache by a path the first did not take. Every one of them
// had a path beneath it that every build read, its package.json among them. So a probe is left out when a
// present path lies beneath it: the descendant can only be there while the directory is, so its own record
// already catches the directory going away, and the probe adds nothing a check could fail on. What remains is
// the same set whichever loader went first. A path is recorded without a trailing slash, since the resolver asks
// for some directories both ways.
func (r *InputRecorder) Inputs() (present []string, absent []string, probed []string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	var existenceOnly []string
	for path, existed := range r.existed {
		if r.written[path] {
			continue
		}
		switch {
		case !existed:
			absent = append(absent, path)
		case !r.depended[path]:
			existenceOnly = append(existenceOnly, path)
		default:
			present = append(present, path)
		}
	}
	// Every present path, probes included, can imply a probe above it.
	everyPresent := append(append([]string{}, present...), existenceOnly...)
	sort.Strings(everyPresent)
	for _, path := range existenceOnly {
		if hasPathBeneath(everyPresent, path) {
			continue
		}
		present = append(present, path)
		probed = append(probed, path)
	}
	sort.Strings(present)
	sort.Strings(absent)
	sort.Strings(probed)
	return present, absent, probed
}

// hasPathBeneath reports whether sorted holds a path inside directory.
func hasPathBeneath(sorted []string, directory string) bool {
	prefix := directory + "/"
	if directory == "/" {
		prefix = "/"
	}
	position := sort.SearchStrings(sorted, prefix)
	return position < len(sorted) && strings.HasPrefix(sorted[position], prefix) && sorted[position] != directory
}

// withoutTrailingSlash is path with any trailing slash removed, the root excepted.
func withoutTrailingSlash(path string) string {
	for len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	return path
}

// recordingFS is a vfs.FS that reports each path it serves to an InputRecorder.
type recordingFS struct {
	vfs.FS
	recorder *InputRecorder
}

// noteAnswer records a path the compiler asked about, by what is on disk rather than by the answer.
//
// A negative answer to a typed question is not absence. Module resolution asks whether
// `./Graph.css` is a directory, gets no, and moves on, and the file is right there. Recording that
// "no" as "absent" made the next check find the file and report "now exists" on an untouched tree,
// four-month-old file, so the cache never hit. FileExists on a directory has the same shape. So a
// negative answer is followed by a stat: anything on disk in any form is recorded as present, with
// its full signature, which also catches a file turning into a directory. Only a path that is
// genuinely not there is recorded absent.
//
// depended says whether the answer depended on more than existence: a read, a listing or a stat.
func (f *recordingFS) noteAnswer(path string, positive bool, depended bool) {
	if !positive {
		positive = f.FS.Stat(tspath.RootedPath(path)) != nil
	}
	f.recorder.note(path, positive, depended)
}

func (f *recordingFS) FileExists(path tspath.RootedFilePath) bool {
	exists := f.FS.FileExists(path)
	f.noteAnswer(path.AsString(), exists, false)
	return exists
}

func (f *recordingFS) ReadFile(path tspath.RootedFilePath) (string, bool) {
	contents, ok := f.FS.ReadFile(path)
	f.noteAnswer(path.AsString(), ok, true)
	return contents, ok
}

func (f *recordingFS) DirectoryExists(path tspath.RootedDirectoryPath) bool {
	exists := f.FS.DirectoryExists(path)
	f.noteAnswer(path.AsString(), exists, false)
	return exists
}

// GetAccessibleEntries is a directory listing, and a listing's answer changes when an entry is added.
// Recording the directory means its mtime is watched, which is what catches a file added beside the
// ones the build read.
func (f *recordingFS) GetAccessibleEntries(path tspath.RootedDirectoryPath) vfs.Entries {
	entries := f.FS.GetAccessibleEntries(path)
	f.noteAnswer(path.AsString(), f.FS.DirectoryExists(path), true)
	return entries
}

func (f *recordingFS) Stat(path tspath.RootedPath) vfs.FileInfo {
	information := f.FS.Stat(path)
	f.noteAnswer(path.AsString(), information != nil, true)
	return information
}

func (f *recordingFS) Realpath(path tspath.RootedPath) tspath.RootedPath {
	resolved := f.FS.Realpath(path)
	// Realpath answers for a missing path too, by returning it unchanged, so existence is asked
	// rather than assumed. A probe recorded as present that is not would fail the record outright.
	f.noteAnswer(path.AsString(), false, false)
	return resolved
}

func (f *recordingFS) WriteFile(path tspath.RootedFilePath, data string) error {
	f.markWritten(path.AsString())
	return f.FS.WriteFile(path, data)
}

func (f *recordingFS) AppendFile(path tspath.RootedFilePath, data string) error {
	f.markWritten(path.AsString())
	return f.FS.AppendFile(path, data)
}

func (f *recordingFS) Remove(path tspath.RootedPath) error {
	f.markWritten(path.AsString())
	return f.FS.Remove(path)
}

func (f *recordingFS) Chtimes(path tspath.RootedPath, accessTime time.Time, modifiedTime time.Time) error {
	f.markWritten(path.AsString())
	return f.FS.Chtimes(path, accessTime, modifiedTime)
}

func (f *recordingFS) markWritten(path string) {
	path = withoutTrailingSlash(path)
	f.recorder.mutex.Lock()
	f.recorder.written[path] = true
	f.recorder.mutex.Unlock()
}

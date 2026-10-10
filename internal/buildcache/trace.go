package buildcache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"syscall"
)

// A traced build is Workshop's (#vt46geg): cmd/traced runs a command under strace -f with ADAMIC_BUILD_TRACE naming
// a directory, and every process under it that builds a product marks the build in the trace and writes it to the
// directory's journal. When the command ends, Settle reads the trace, gives each build the reads of its process and
// the processes it started, up to the build's end, and records the read set that becomes the product's key. Only a
// settled build is placed under its key and published, so only a traced build publishes.
//
// A marker is a failed access of /adamic-trace/<event>/<id>: it prints in the trace beside the thread that made it and
// changes nothing else. Keying (hashing read sets and declared files, go list, go env, a tool's report, git ls-files)
// runs locked to one thread between key-begin and key-end, so what a key reads is never counted as what a build reads.
const traceMarker = "/adamic-trace/"

// traceDirectory is ADAMIC_BUILD_TRACE: the directory of the trace this process runs under, or "" when untraced.
func traceDirectory() string {
	return os.Getenv("ADAMIC_BUILD_TRACE")
}

func mark(event, id string) {
	if traceDirectory() != "" {
		_ = syscall.Access(traceMarker+event+"/"+id, 0)
	}
}

// keying runs compute on one thread between a key-begin and a key-end marker; the processes it starts are its own.
func keying(compute func()) {
	if traceDirectory() == "" {
		compute()
		return
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	mark("key-begin", "-")
	defer mark("key-end", "-")
	compute()
}

var builds atomic.Int64

// traced runs a product's build between build-begin and build-end markers naming it, and returns the build's id.
func traced(nameKey string, build func() error) (string, error) {
	id := fmt.Sprintf("%s.%d.%d", nameKey[:16], os.Getpid(), builds.Add(1))
	mark("build-begin", id)
	defer mark("build-end", id)
	return id, build()
}

// A journalEntry is one line of a trace's journal: a process that keys products (its executable and the package and
// tags it was built from, Settle's recipe), a product it found (by name key and directory, so reads under that
// directory are reads of that product), or a product it built inside a marked build.
type journalEntry struct {
	Event      string   `json:"event"`
	Process    int      `json:"process"`
	Executable string   `json:"executable,omitempty"`
	Package    string   `json:"package,omitempty"`
	Settings   []string `json:"settings,omitempty"`
	Build      string   `json:"build,omitempty"`
	NameKey    string   `json:"nameKey,omitempty"`
	Name       string   `json:"name,omitempty"`
	Directory  string   `json:"directory,omitempty"`
	Inputs     string   `json:"inputs,omitempty"`
	Declared   []string `json:"declared,omitempty"`
	Root       string   `json:"root,omitempty"`
}

var journalProcess sync.Once

// journal appends entry to the trace's journal, after this process's own entry. One write per line, appended, so
// processes sharing a journal never interleave inside a line. A journal that can't be written is a lost trace: Settle
// refuses what it can't place.
func journal(entry journalEntry) {
	directory := traceDirectory()
	if directory == "" {
		return
	}
	journalProcess.Do(func() {
		own := journalEntry{Event: "process", Process: os.Getpid()}
		own.Executable, _ = os.Executable()
		if info, ok := debug.ReadBuildInfo(); ok {
			own.Package = info.Path
			for _, setting := range info.Settings {
				own.Settings = append(own.Settings, setting.Key+"="+setting.Value)
			}
		}
		appendJournal(directory, own)
	})
	entry.Process = os.Getpid()
	appendJournal(directory, entry)
}

func appendJournal(directory string, entry journalEntry) {
	line, err := json.Marshal(entry)
	if err != nil {
		return
	}
	file, err := os.OpenFile(filepath.Join(directory, "journal.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	file.Write(append(line, '\n'))
	file.Close()
}

// pendingDirectory holds the products this trace built and hasn't settled yet, by name key: every process under the
// same trace finds them there, and nothing outside it ever does.
func pendingDirectory(cache string) string {
	return filepath.Join(cache, "pending", filepath.Base(traceDirectory()))
}

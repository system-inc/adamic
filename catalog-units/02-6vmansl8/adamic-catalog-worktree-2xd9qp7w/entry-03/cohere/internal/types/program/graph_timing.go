package program

import (
	"sync/atomic"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// GraphTiming is what building the graph cost, by part, collected when Options.Timing is set.
//
// The graph phase was one number, `graph built in 547ms`, and nobody knew what it was made of
// (#cazsft3). The levers proposed for it were guesses: parallel directory enumeration, a resolution
// cache, a declaration pack. Each is worth something only if its part is a large share, and that is
// what this measures rather than reasons about.
//
// Two kinds of number, and they must not be read as one:
//
//   - Wall: Config, Program and Verify are each one stretch of the build's own goroutine, and they add
//     up to most of the phase.
//   - Summed: the compiler loads files on many goroutines at once, so the time spent inside its loads
//     and inside the disk is summed across all of them and runs far past Program's wall. It says
//     where the loaders' work goes, not how long anyone waited.
//
// Parse and module resolution are not separable from here. Both run inside the compiler's file
// loader, interleaved per file, behind one call. What can be measured from outside is each load,
// which is a read and a parse, and the disk beneath it; resolution and the loader's own bookkeeping
// are the part of Program's wall those leave unexplained.
type GraphTiming struct {
	// Builds is how many times the program was built: two when the first build's roots moved under it
	// and Build built again. Every duration below covers all of them.
	Builds int

	// Config is parsing the tsconfig, which includes enumerating the directories its `include`
	// patterns reach. That enumeration is the directory walk a parallel readDirectory would speed up.
	Config time.Duration

	// Program is compiler.NewProgram: every file read, parsed and its imports resolved.
	Program time.Duration

	// Verify is matching the config's files to the program's, and asking the disk about any missing.
	Verify time.Duration

	// SourceFileLoads and SourceFileSummed are the compiler host's GetSourceFile calls, each a read
	// and a parse, summed across the loaders.
	SourceFileLoads  int64
	SourceFileSummed time.Duration

	// ConfigDisk and ProgramDisk are the real filesystem calls made while the config was parsed and
	// while the program was built: beneath the stat and readdir cache, so a call it answered is not
	// here, and above the content pack, so a file the pack served is here at the cost of serving it.
	ConfigDisk  DiskTiming
	ProgramDisk DiskTiming

	// PackServed and PackRead are the content pack's own count of the reads above: files it served from
	// its mapping, and files it had to read from disk. Both zero when the build had no pack. A read the
	// pack served costs a copy, not an open, so the two must be told apart to read Reads' time.
	PackServed int
	PackRead   int

	// CheckedAnswers is how many of the existence checks and stats above were answered from the run cache's
	// check (Options.CheckedStats) rather than by the disk, across config and program alike.
	CheckedAnswers int64
}

// DiskTiming is filesystem calls by kind, each a count and the time summed across every goroutine
// that made one.
type DiskTiming struct {
	Reads     DiskCalls // ReadFile
	Existence DiskCalls // FileExists and DirectoryExists
	Stats     DiskCalls // Stat
	Listings  DiskCalls // GetAccessibleEntries, a directory listing
	Realpaths DiskCalls // Realpath
}

// DiskCalls is a count of one kind of call and their summed duration.
type DiskCalls struct {
	Count  int64
	Summed time.Duration
}

// Summed is every call's duration, added up.
func (d DiskTiming) Summed() time.Duration {
	return d.Reads.Summed + d.Existence.Summed + d.Stats.Summed + d.Listings.Summed + d.Realpaths.Summed
}

// diskCounters accumulates calls from many goroutines at once. Atomic rather than locked, because a
// lock here would serialize the loaders and measure itself.
type diskCounters struct {
	counts    [diskCallKinds]atomic.Int64
	durations [diskCallKinds]atomic.Int64
}

const (
	diskRead = iota
	diskExistence
	diskStat
	diskListing
	diskRealpath
	diskCallKinds
)

func (c *diskCounters) add(kind int, started time.Time) {
	c.counts[kind].Add(1)
	c.durations[kind].Add(int64(time.Since(started)))
}

func (c *diskCounters) snapshot() DiskTiming {
	calls := func(kind int) DiskCalls {
		return DiskCalls{Count: c.counts[kind].Load(), Summed: time.Duration(c.durations[kind].Load())}
	}
	return DiskTiming{
		Reads:     calls(diskRead),
		Existence: calls(diskExistence),
		Stats:     calls(diskStat),
		Listings:  calls(diskListing),
		Realpaths: calls(diskRealpath),
	}
}

func (a DiskTiming) plus(b DiskTiming) DiskTiming {
	sum := func(x DiskCalls, y DiskCalls) DiskCalls {
		return DiskCalls{Count: x.Count + y.Count, Summed: x.Summed + y.Summed}
	}
	return DiskTiming{
		Reads:     sum(a.Reads, b.Reads),
		Existence: sum(a.Existence, b.Existence),
		Stats:     sum(a.Stats, b.Stats),
		Listings:  sum(a.Listings, b.Listings),
		Realpaths: sum(a.Realpaths, b.Realpaths),
	}
}

// timingFS times the calls beneath it into whichever counters are current. The build switches them
// from the config's to the program's between the two, which is safe because both run on the build's
// own goroutine in order: no config call is still in flight when the program starts.
type timingFS struct {
	vfs.FS
	current atomic.Pointer[diskCounters]
}

func (f *timingFS) start() (*diskCounters, time.Time) {
	return f.current.Load(), time.Now()
}

func (f *timingFS) ReadFile(path tspath.RootedFilePath) (string, bool) {
	counters, started := f.start()
	contents, ok := f.FS.ReadFile(path)
	counters.add(diskRead, started)
	return contents, ok
}

func (f *timingFS) FileExists(path tspath.RootedFilePath) bool {
	counters, started := f.start()
	exists := f.FS.FileExists(path)
	counters.add(diskExistence, started)
	return exists
}

func (f *timingFS) DirectoryExists(path tspath.RootedDirectoryPath) bool {
	counters, started := f.start()
	exists := f.FS.DirectoryExists(path)
	counters.add(diskExistence, started)
	return exists
}

func (f *timingFS) Stat(path tspath.RootedPath) vfs.FileInfo {
	counters, started := f.start()
	information := f.FS.Stat(path)
	counters.add(diskStat, started)
	return information
}

func (f *timingFS) GetAccessibleEntries(path tspath.RootedDirectoryPath) vfs.Entries {
	counters, started := f.start()
	entries := f.FS.GetAccessibleEntries(path)
	counters.add(diskListing, started)
	return entries
}

func (f *timingFS) Realpath(path tspath.RootedPath) tspath.RootedPath {
	counters, started := f.start()
	resolved := f.FS.Realpath(path)
	counters.add(diskRealpath, started)
	return resolved
}

// timingHost times the compiler host's GetSourceFile, which is how the program loads every file: a
// read and a parse. Everything else passes through.
type timingHost struct {
	compiler.CompilerHost
	loads  atomic.Int64
	summed atomic.Int64
}

func (h *timingHost) GetSourceFile(options ast.SourceFileParseOptions) *ast.SourceFile {
	started := time.Now()
	sourceFile := h.CompilerHost.GetSourceFile(options)
	h.loads.Add(1)
	h.summed.Add(int64(time.Since(started)))
	return sourceFile
}

package rule_testing

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"sync"

	"github.com/system-inc/cohere/internal/types/program"
)

// Building a type graph is the most expensive thing this harness does, and a build is memoized on
// exactly what determines its result: the fixture file set and the subject's name. Two calls with the
// same bytes cannot produce different graphs, which is what makes sharing one safe rather than merely
// fast.
//
// The memo is bounded, because unbounded it was the largest memory cost in the suite (#7nx3z1n): measured
// at GOMAXPROCS=3 it peaked at 69 GB in registry, 31 GB in rules/core and 29 GB in rules/typescript, past
// a 7 GB CI runner, which then swapped into the test timeout. Most of it held nothing a later test would
// ask for. Counted per package, rules/core asked 8,269 times for 6,923 distinct fixtures, and registry
// 20,046 times for 16,353, so nearly every entry was used once and kept forever.
//
// And every entry was heavy for a reason that had nothing to do with its fixture: each one-file program
// parsed the 58 lib files its ES2022 target names, 3.25ms and 2.5 MB of a build against 0.1ms for the
// fixture itself. So every build here, cached or not, takes its lib files from one shared parse
// (libraryParses, #0rj7wvg), and an entry is only what its fixture adds.
//
// Together, measured interleaved against the unbounded memo at GOMAXPROCS=3: every package peaks under
// 1.6 GB and runs no slower, registry 69 GB and 41s to 372 MB and 10s, rules/core 30 GB and 18.6s to
// 99 MB and 5.7s. The corpus walkers in `high_level_intermediate_representation` were the case a bound
// alone could not serve: thirty-seven walks of one pinned corpus ask 14,730 times for 1,529 fixtures,
// but each recurs only after the whole corpus has gone by, which no affordable LRU reaches. Bounded
// without the shared parse they ran 66s to about 93s; with it, 76.5s to 69.4s at 117 MB against 9.9 GB,
// because a rebuild now costs the fixture and not its libs.
//
// What is deliberately NOT shared is the checker. Upstream owns a checker per worker, keyed by the
// file's position in the program's file list, and it is thread-affine: reaching one from two
// goroutines produced `fatal error: concurrent map read and map write` inside `core.LinkStore.Get`,
// which is a runtime fatal rather than a recoverable panic. `checker_concurrency_test.go` records
// that in full. Callers keep going through `Graph.CheckerForFile`, which hands out the exclusive
// form and a release, so each test still acquires and releases its own. The cache holds the built
// graph; the checker lease stays per-test.

// programCacheEntry is one memoized build, with the error it produced if it failed.
//
// The error is cached alongside the graph so a fixture that cannot build fails the same way for
// every caller instead of being rebuilt on each attempt.
type programCacheEntry struct {
	once      sync.Once
	graph     *program.Graph
	buildErr  error
	directory string

	// key is the entry's key in programCache, and recency its place in programCacheRecency.
	key     string
	recency *list.Element
}

var (
	// libraryParses is the bundled lib files, parsed once for every program this test binary builds,
	// cached or not. Each fixture otherwise parsed its target's 58 lib files again, which was most of a
	// build's time and nearly all of what a built program kept (#7nx3z1n, #0rj7wvg).
	libraryParses = program.NewLibraryParses()

	programCacheMutex sync.Mutex
	programCache      = map[string]*programCacheEntry{}

	// programCacheRecency orders the entries most recently used first, so the least recently used is
	// the one evicted when the cache is full.
	programCacheRecency = list.New()

	// programCacheCapacity is how many built programs the cache holds at once.
	programCacheCapacity = defaultProgramCacheCapacity

	// programCacheBuilds counts builds, so each one gets a root of its own. A key evicted and built
	// again must not share a root with the first build, which a test may still be holding.
	programCacheBuilds int
)

// programCacheKey hashes everything that can change what a build produces.
//
// The file names are sorted so a caller's map iteration order cannot split one program into two
// cache entries, and each field is length-prefixed so that no combination of names and contents can
// be spelled two ways and collide.
func programCacheKey(files map[string]string, subjectFileName string, verbatim bool) string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	hash := sha256.New()
	fmt.Fprintf(hash, "subject:%d:%s\n", len(subjectFileName), subjectFileName)
	fmt.Fprintf(hash, "verbatim:%t\n", verbatim)
	for _, name := range names {
		fmt.Fprintf(hash, "file:%d:%s:%d:%s\n", len(name), name, len(files[name]), files[name])
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// buildCachedProgram returns the graph for a fixture set, building it at most once per distinct set.
//
// The returned directory is the fixture's root in its memory filesystem, for a caller that needs to name a
// path inside it. Nothing on disk is there.
func buildCachedProgram(files map[string]string, subjectFileName string, verbatim bool) (*program.Graph, string, error) {
	key := programCacheKey(files, subjectFileName, verbatim)

	programCacheMutex.Lock()
	entry, found := programCache[key]
	if found {
		programCacheRecency.MoveToFront(entry.recency)
	} else {
		programCacheBuilds++
		entry = &programCacheEntry{key: key, directory: fmt.Sprintf("%s-%d", key, programCacheBuilds)}
		entry.recency = programCacheRecency.PushFront(entry)
		programCache[key] = entry
		evictBeyondCapacity()
	}
	programCacheMutex.Unlock()

	// Outside the map lock, so one slow build does not stall every other fixture, and `sync.Once`
	// still guarantees exactly one build per key even when several tests race for the same one.
	entry.once.Do(func() {
		entry.graph, entry.directory, entry.buildErr = buildProgramInto(entry.directory, files, subjectFileName, verbatim)
	})

	return entry.graph, entry.directory, entry.buildErr
}

// defaultProgramCacheCapacity is how many built programs a test binary keeps. Sixteen, chosen by
// measurement before the shared lib parse (#7nx3z1n), when an entry still held its libs: every rule
// package and registry peaked under 1 GB at sixteen and ran no slower than unbounded, while rules/nexus,
// whose fixtures are the largest, reached 2.1 GB at sixty-four, and the corpus walkers ran no faster at
// 128 than at sixteen. With the shared parse an entry is far smaller, rules/nexus peaking at 151 MB, so
// sixteen now has room; it was not raised, because no package measured slower at it.
const defaultProgramCacheCapacity = 16

// evictBeyondCapacity drops the least recently used entries until the cache is within its capacity.
// The caller holds programCacheMutex. An evicted graph lives on in any test still holding it, and is
// collected when the last one lets go, with the memory filesystem it reads.
func evictBeyondCapacity() {
	for len(programCache) > programCacheCapacity {
		oldest := programCacheRecency.Back()
		evicted := oldest.Value.(*programCacheEntry)
		programCacheRecency.Remove(oldest)
		delete(programCache, evicted.key)
	}
}

// memoryFixtureRoot is the directory every cached fixture's own root sits under. Only the fixture's memory
// filesystem holds it, so no path a cached build reads is on disk.
const memoryFixtureRoot = "/cohere-fixture"

// buildProgramInto builds one fixture set from memory, rooted at a directory of its own named
// directoryName under memoryFixtureRoot, and returns that root.
//
// Nothing is written to disk. Each cached build used to write a directory under the cache root, about
// 16,000 for a registry run, and three costs came with them: the writes, deleting them at exit (3.7s of
// registry's 7.6s), and the module resolution's stats and opens around them, which went into the test log
// Go replays on every cached run (about 420,000 lines for registry, 6.5s to say "cached") (#nxgt2ca).
func buildProgramInto(directoryName string, files map[string]string, subjectFileName string, verbatim bool) (*program.Graph, string, error) {
	directory := memoryFixtureRoot + "/" + directoryName
	contents := make(map[string]string, len(files)+1)
	for name, text := range files {
		if !verbatim {
			text = FixtureText(text)
		}
		contents[directory+"/"+filepath.ToSlash(name)] = text
	}
	configPath := directory + "/tsconfig.json"
	contents[configPath] = defaultTsConfig

	graph, err := program.Build(program.Options{
		ConfigFileName:   configPath,
		CurrentDirectory: directory,
		LibraryParses:    libraryParses,
		FileSystem:       program.NewMemoryFS(contents),
		// One checker rather than several: a fixture is one file, and the parallel path adds
		// scheduling nondeterminism to a test whose whole value is being deterministic.
		SingleThreaded: true,
	})
	if err != nil {
		return nil, directory, fmt.Errorf("building the type graph: %w", err)
	}
	return graph, directory, nil
}

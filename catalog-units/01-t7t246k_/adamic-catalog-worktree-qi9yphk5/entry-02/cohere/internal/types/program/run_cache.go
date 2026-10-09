package program

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// The run cache: replay a whole run's output without building anything, when every input it
// depended on is provably unchanged.
//
// This is a different cache from the lint findings cache beside it, and the difference is the whole
// reason it exists. That cache was measured at about 9% and left unwired because it still walked the
// tree: it could only skip map lookups the one-traversal design had already made nearly free. This
// one skips the walk AND the graph build, which is roughly 0.45s on the ahra tree before any rule
// runs. Its prize is the entire run, so the earlier verdict says nothing about it.
//
// # The rule it has to keep
//
// A stale run cache that replays a clean verdict over a tree that is no longer clean is the worst
// thing this tool could do, and it would look exactly like success. So the cache answers "hit" only
// when it can prove every input is unchanged, and any doubt at all is a miss: an unreadable
// manifest, a format version it does not recognize, a single entry that does not match. A miss costs
// one ordinary run. A false hit costs a lie.
//
// # The input it cannot see, and how it sees it anyway
//
// The program's file list only exists after the graph is built, which is the cost this cache exists
// to skip. So a full run records the list and the next run re-stats what was recorded. That is sound
// for a file edited or deleted. It is blind to a file ADDED: a new source file under an include glob
// is an input the recorded list has never heard of, and re-statting the old list finds nothing wrong.
//
// Directories close that gap. Adding or removing an entry changes the mtime of the directory that
// holds it, so the manifest records every directory the program draws from and re-stats those too.
// Every directory, not the include roots: verified on macOS, a file added in a subdirectory changes
// that subdirectory's mtime and leaves its parent's alone, so watching only the roots would miss
// anything added one level down. On the ahra tree that is 1,725 directories beside 10,360 files, and
// statting all 12,085 costs about 13ms across eight workers.
//
// # What "unchanged" means for one entry
//
// Size and modification time, at nanosecond resolution. A file whose bytes changed and whose mtime
// did not is the case this cannot see, and it requires something deliberately resetting the mtime:
// `touch -r`, some archive extractors, a tool that preserves timestamps on copy. Content hashing every
// file would close it at a cost of reading 110 MB per run, which is the run this cache exists to
// avoid. The trade is stated here so nobody discovers it by surprise; Record also hashes each file's
// content so a later change of policy has the data to tighten this without a format change.
type RunCache struct {
	// Version is the manifest format. A manifest written by a different format is a miss, never a
	// best-effort read: the fields would be read from the wrong places and still parse.
	Version int

	// Key covers everything about the run that is not a file on disk: the flags, the binary's
	// identity, the working directory. Two runs that differ in any of those are different runs even
	// over identical files, and must not replay each other.
	Key string

	// KeyParts is Key's parts, so a run under another key can say which part moved. See RunCacheKeyParts.
	KeyParts RunCacheKeyParts

	// Inputs is every file and directory the run depended on, with the signature it had.
	Inputs []RunCacheInput

	// Output is exactly what the run printed to stdout, replayed byte for byte on a hit.
	Output []byte

	// Errors is what it printed to stderr. Kept apart rather than interleaved, so each stream replays
	// to the stream it was written to.
	Errors []byte

	// RecordedUnixNanoseconds is when the replayed run happened. A replay prints that run's report,
	// durations included, so it has to be able to say when those numbers are from.
	RecordedUnixNanoseconds int64

	// ExitCode is the process exit code to return on a hit. A cache that replayed the output and
	// exited zero on a failing tree would pass CI while printing a failure.
	ExitCode int

	// Summary is the run's summary as the command encodes it, opaque here. The footer a run ends with
	// carries that run's time, so it is never in Output; a replay renders its own from this, with the
	// replay's time, rather than reprinting a recorded one that would claim the recorded run's.
	Summary []byte
}

// RunCacheInput is one input's signature.
type RunCacheInput struct {
	Path string

	// Directory is true for an entry recorded to catch files added beneath it. Kept explicit
	// rather than inferred, because a path that was a directory and is now a file is a change.
	Directory bool

	// Exists is false for an input that was absent when recorded and must stay absent. A config the
	// run looked for and did not find is still an input: creating it changes the run.
	Exists bool

	Size                int64
	ModifiedNanoseconds int64

	// ChangedNanoseconds and Inode catch what the modification time cannot: different bytes of the same
	// size under a restored modification time, as cp -p, rsync -t and touch -r leave a file. Nothing in
	// userland can set a change time, and a file replaced by a rename has a new inode. Zero where the
	// platform's stat does not give them.
	ChangedNanoseconds int64
	Inode              uint64

	// ExistenceOnly is set on a directory the run only asked whether it exists (see InputRecorder.Inputs),
	// which matches while it is still the same directory, by inode, whatever was added to it since. Module
	// resolution asks it of every directory from the project up to the root, and on a busy machine /tmp and
	// the like change many times a second: checked at full signature, they cost a replay on a tree nothing
	// had touched (#q51f02a). A package.json or node_modules appearing in one is its own absent input.
	ExistenceOnly bool
}

// runCacheVersion is bumped whenever the manifest's meaning changes, not only its shape.
//
// 2: Output no longer carries the lines that describe the recorded invocation (graph built in, types and
// lint durations, the phases and total lines). A version-1 manifest has them, and replaying one would
// print a graph build that did not happen.
//
// 3: Output carries a tag on lines a replay prints with where they came from, the fix line among them.
// A version-2 manifest has none, and would replay "fix: ..." as though the fix phase had just run.
//
// 4: Inputs carry their change time and inode. A version-3 input has neither, and checked against a
// stat that does, every one would read as changed.
//
// 5: A run ends with a footer that is not in Output, and Summary carries what a replay renders it from.
// A version-4 manifest has the old phase account in Output and no Summary, and would replay with no
// footer at all.
const runCacheVersion = 5

// ErrRunCacheMiss is the one answer a check gives when it cannot prove a hit. Callers treat every
// error from Check as a miss and run normally; this exists so a test can tell a clean miss
// from an unexpected failure.
var ErrRunCacheMiss = errors.New("run cache miss")

// RunCacheKey hashes what identifies a run apart from its files. It is RunCacheKeyPartsOf's key.
func RunCacheKey(arguments []string, workingDirectory string, facts ...string) (string, error) {
	parts, err := RunCacheKeyPartsOf(arguments, workingDirectory, facts...)
	if err != nil {
		return "", err
	}
	return parts.Key(), nil
}

// RunCacheKeyParts is what identifies a run apart from its files, in four parts hashed apart, so a run that
// misses can say which one moved rather than that something did (#547dhjz).
type RunCacheKeyParts struct {
	// Binary is the running binary: its resolved path, size and modification time, and the format and
	// platform it reads. Keyed by path, size and time rather than by hashing its bytes: it is about 70 MB,
	// reading it every run would spend the budget this cache exists to save, and a rebuilt binary has a
	// new modification time, which is all this needs.
	Binary string

	// Directory is the working directory the caller resolved, the project root's target.
	Directory string

	// Arguments is the command line.
	Arguments string

	// Facts is anything else the run's output depends on that is not a file the build reads: config paths
	// the command resolved, and results it computes from sources the build never sees. Order matters, and
	// an empty fact still counts.
	Facts string
}

// RunCacheKeyPartsOf hashes each part of a run's identity.
func RunCacheKeyPartsOf(arguments []string, workingDirectory string, facts ...string) (RunCacheKeyParts, error) {
	executable, err := os.Executable()
	if err != nil {
		return RunCacheKeyParts{}, fmt.Errorf("locating the running binary: %w", err)
	}
	// Resolved, as the callers resolve the directory, so a binary reached through a link keys as itself (#547dhjz).
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		executable = resolved
	}
	information, err := os.Stat(executable)
	if err != nil {
		return RunCacheKeyParts{}, fmt.Errorf("reading the running binary: %w", err)
	}
	part := func(values ...string) string {
		hash := sha256.New()
		for _, value := range values {
			hash.Write([]byte(value))
			hash.Write([]byte{0})
		}
		return fmt.Sprintf("%x", hash.Sum(nil))
	}
	return RunCacheKeyParts{
		Binary: part(fmt.Sprintf("version %d", runCacheVersion), executable, fmt.Sprintf("%d", information.Size()),
			fmt.Sprintf("%d", information.ModTime().UnixNano()), runtime.GOOS, runtime.GOARCH),
		Directory: part(workingDirectory),
		Arguments: part(append([]string{fmt.Sprintf("%d", len(arguments))}, arguments...)...),
		Facts:     part(append([]string{fmt.Sprintf("%d facts", len(facts))}, facts...)...),
	}, nil
}

// Key is the parts as one key.
func (parts RunCacheKeyParts) Key() string {
	sum := sha256.Sum256([]byte(parts.Binary + "\x00" + parts.Directory + "\x00" + parts.Arguments + "\x00" + parts.Facts))
	return fmt.Sprintf("%x", sum)
}

// Changed says which parts differ from recorded, in words, or that the record kept no parts to compare, as
// one written before they were kept.
func (parts RunCacheKeyParts) Changed(recorded RunCacheKeyParts) string {
	if recorded == (RunCacheKeyParts{}) {
		return "it was recorded without its key's parts, so which one moved cannot be said"
	}
	var changed []string
	if parts.Binary != recorded.Binary {
		changed = append(changed, "the cohere binary")
	}
	if parts.Directory != recorded.Directory {
		changed = append(changed, "the project root")
	}
	if parts.Arguments != recorded.Arguments {
		changed = append(changed, "the arguments")
	}
	if parts.Facts != recorded.Facts {
		changed = append(changed, "the configuration (the tsconfig or lint config, their extends chains, the rules, or the output's color)")
	}
	if len(changed) == 0 {
		return "no part changed, though the key did"
	}
	return strings.Join(changed, " and ") + " changed"
}

// RecordRunCache captures a run that just finished.
//
// files is every path the run depended on that existed, and directories among them are fine: whether
// an entry is a file or a directory is read from the disk here, never inferred from which list it
// arrived in. An earlier version stamped every entry of files as a file, so a directory the compiler
// had probed was recorded as one and every check then reported it "changed between file and
// directory". The cache never hit, which is the most deceptive way for a cache to fail: it runs, it
// misses, and it looks merely cold. The probe that found it was the untouched-tree control.
//
// The directory holding each file is added here too, so a caller cannot forget them and quietly ship
// a cache blind to added files. extraDirectories adds more, and absent lists paths the run looked for
// and did not find, which must stay absent for a hit. probed lists present paths the run only asked
// whether they exist; a directory among them that holds no file the run read and is not among
// extraDirectories is recorded ExistenceOnly.
//
// readSince is when the run began reading its inputs. An input changed after it may have been read before
// the change, so the stat taken here would sign the new bytes against a verdict computed from the old ones,
// and the next run would replay that verdict over a tree it does not describe: a file saved in an editor
// while cohere runs is exactly this. Such a run is not recorded. The zero time checks nothing.
func RecordRunCache(key string, files []string, extraDirectories []string, absent []string, probed []string, output []byte, exitCode int, readSince time.Time) (*RunCache, error) {
	cache := &RunCache{Version: runCacheVersion, Key: key, Output: output, ExitCode: exitCode}

	// One entry per path. A path can arrive as a file read, a directory listed, and the parent of
	// something else, and statting it three times would spend the budget on duplicates.
	seen := map[string]bool{}
	var present []string
	sortedFiles := append([]string(nil), files...)
	sort.Strings(sortedFiles)
	for _, file := range sortedFiles {
		if !seen[file] {
			seen[file] = true
			present = append(present, file)
		}
	}
	inputs, err := signaturesOf(present)
	if err != nil {
		return nil, err
	}

	// The directory holding each file, so a file added beside one the run read is noticed. Only a file's:
	// a directory the run listed is an input by its own signature, whose change time and inode already catch
	// it being replaced, so its parent adds nothing. Recording the parent of the project's own directory made
	// every replay depend on whatever else lives beside the project, ~/Projects on a developer's machine and
	// the benchmark's work directory, where something is created all day (#r9jevk9).
	directorySet := map[string]struct{}{}
	for _, input := range inputs {
		if !input.Directory {
			directorySet[filepath.Dir(input.Path)] = struct{}{}
		}
	}
	for _, directory := range extraDirectories {
		directorySet[directory] = struct{}{}
	}
	directories := make([]string, 0, len(directorySet))
	for directory := range directorySet {
		if !seen[directory] {
			seen[directory] = true
			directories = append(directories, directory)
		}
	}
	sort.Strings(directories)
	directoryInputs, err := signaturesOf(directories)
	if err != nil {
		return nil, err
	}
	cache.Inputs = append(inputs, directoryInputs...)

	// A directory only probed, holding nothing the run read and named by no caller, is kept for its existence.
	probedSet := make(map[string]struct{}, len(probed))
	for _, path := range probed {
		probedSet[path] = struct{}{}
	}
	for index, input := range cache.Inputs {
		_, onlyProbed := probedSet[input.Path]
		_, needed := directorySet[input.Path]
		if input.Directory && onlyProbed && !needed {
			cache.Inputs[index].ExistenceOnly = true
		}
	}
	if !readSince.IsZero() {
		// Files, and the directories holding them, where a file added mid-run beside the ones read would be
		// missed. A directory only probed on the way to a node_modules or a package.json is left out: such
		// ancestors sit above the project, other processes move them all the time (a temporary directory
		// never stops), and checking them would keep most runs from ever being recorded.
		holdsARead := map[string]bool{}
		for _, input := range cache.Inputs {
			if !input.Directory {
				holdsARead[filepath.Dir(input.Path)] = true
			}
		}
		for _, input := range cache.Inputs {
			if input.ExistenceOnly || input.Directory && !holdsARead[input.Path] {
				continue
			}
			if max(input.ModifiedNanoseconds, input.ChangedNanoseconds) > readSince.UnixNano() {
				return nil, fmt.Errorf("%s changed after the run began reading, so what the run saw of it is not known", input.Path)
			}
		}
	}

	for _, path := range absent {
		if seen[path] {
			// Probed absent once and found present another time is a contradiction within one run.
			// The present reading wins, since the run did read it.
			continue
		}
		seen[path] = true
		cache.Inputs = append(cache.Inputs, RunCacheInput{Path: path, Exists: false})
	}

	return cache, nil
}

// signaturesOf stats every path in parallel, as Check does, each into its own slot, so the order and the
// first error are the ones a serial pass would give. About 17,000 stats on ahra: 64 to 79ms one at a time,
// under 20ms across workers (#a66sfmh), on the path between the report printing and the process exiting.
func signaturesOf(paths []string) ([]RunCacheInput, error) {
	inputs := make([]RunCacheInput, len(paths))
	failures := make([]error, len(paths))
	workers := min(runtime.NumCPU(), 8)
	var group sync.WaitGroup
	for worker := range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := worker; index < len(paths); index += workers {
				inputs[index], failures[index] = signatureOf(paths[index])
			}
		}()
	}
	group.Wait()
	for _, err := range failures {
		if err != nil {
			return nil, err
		}
	}
	return inputs, nil
}

// signatureOf stats one input. A path that vanished between the run and the record is an error rather
// than an absent entry: recording it as absent would make the next run's absence look like a match,
// and the run that just finished did read it.
func signatureOf(path string) (RunCacheInput, error) {
	information, err := os.Stat(path)
	if err != nil {
		return RunCacheInput{}, fmt.Errorf("recording %s: %w", path, err)
	}
	changed, inode := changeTimeAndInode(path, information)
	return RunCacheInput{
		Path:                path,
		Directory:           information.IsDir(),
		Exists:              true,
		Size:                information.Size(),
		ModifiedNanoseconds: information.ModTime().UnixNano(),
		ChangedNanoseconds:  changed,
		Inode:               inode,
	}, nil
}

// Check reports whether a stored run can be replayed, returning nil only on a provable hit.
//
// It returns the stored run only when the key matches and every input still has the signature it
// was recorded with. The first mismatch is the answer; there is no partial hit, because a run's
// verdict is a property of all its inputs together.
func (c *RunCache) Check(key string) error {
	return c.CheckNoting(key, nil)
}

// CheckNoting is Check, noting in snapshot what it found at every path it statted, so the build and the content
// pack can use stats this run already took rather than take each again (#kdee854). A nil snapshot notes nothing.
func (c *RunCache) CheckNoting(key string, snapshot *StatSnapshot) error {
	if c == nil {
		return fmt.Errorf("%w: no cache", ErrRunCacheMiss)
	}
	if c.Version != runCacheVersion {
		return fmt.Errorf("%w: format version %d, this build reads %d", ErrRunCacheMiss, c.Version, runCacheVersion)
	}
	if c.Key != key {
		return fmt.Errorf("%w: the binary, flags or directory changed", ErrRunCacheMiss)
	}
	if len(c.Inputs) == 0 {
		// A manifest with no inputs would match any tree. A run always depends on something, so
		// an empty list is a manifest that was written wrong, not a run that read nothing.
		return fmt.Errorf("%w: the manifest records no inputs", ErrRunCacheMiss)
	}
	return c.changedInput(snapshot)
}

// ChangedInput reports the first recorded input whose signature no longer matches the disk, nil when every
// one still does. It is Check without the key, so --cache-dump can say which input keeps a run from
// replaying rather than leaving a miss silent (#r9jevk9).
//
// The input named is the earliest recorded one that moved, whichever worker found it, so a run that misses
// says the same thing every time rather than the first thing a worker happened to reach (#547dhjz).
func (c *RunCache) ChangedInput() error {
	return c.changedInput(nil)
}

// changedInput is ChangedInput, noting what each stat found in snapshot when it is not nil.
func (c *RunCache) changedInput(snapshot *StatSnapshot) error {
	workers := min(runtime.NumCPU(), 8)
	var mutex sync.Mutex
	var mismatch error
	earliest := len(c.Inputs)
	var waitGroup sync.WaitGroup
	next := make(chan int, 256)

	for range workers {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for index := range next {
				if err := c.Inputs[index].stillMatches(snapshot); err != nil {
					mutex.Lock()
					if index < earliest {
						earliest, mismatch = index, err
					}
					mutex.Unlock()
				}
			}
		}()
	}
	for index := range c.Inputs {
		next <- index
	}
	close(next)
	waitGroup.Wait()

	return mismatch
}

// stillMatches compares one input against the filesystem now, noting what the stat found in snapshot, matching or
// not: the stat is current either way.
func (input RunCacheInput) stillMatches(snapshot *StatSnapshot) error {
	information, err := os.Stat(input.Path)
	if snapshot != nil {
		snapshot.note(input.Path, information, err)
	}
	if !input.Exists {
		if err == nil {
			return fmt.Errorf("%w: %s now exists", ErrRunCacheMiss, input.Path)
		}
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		// Neither present nor provably absent, for example a permission error. Doubt is a miss.
		return fmt.Errorf("%w: %s cannot be checked: %v", ErrRunCacheMiss, input.Path, err)
	}
	if err != nil {
		return fmt.Errorf("%w: %s: %v", ErrRunCacheMiss, input.Path, err)
	}
	if information.IsDir() != input.Directory {
		return fmt.Errorf("%w: %s changed between file and directory", ErrRunCacheMiss, input.Path)
	}
	if input.ExistenceOnly {
		// Still there and still the same directory, which is all the run asked of it.
		if _, inode := changeTimeAndInode(input.Path, information); inode != input.Inode {
			return fmt.Errorf("%w: %s was replaced", ErrRunCacheMiss, input.Path)
		}
		return nil
	}
	// A directory's size is filesystem bookkeeping and moves with its entries on some filesystems,
	// so only its modification time decides. A file needs both.
	if !input.Directory && information.Size() != input.Size {
		return fmt.Errorf("%w: %s changed size", ErrRunCacheMiss, input.Path)
	}
	if information.ModTime().UnixNano() != input.ModifiedNanoseconds {
		return fmt.Errorf("%w: %s was modified", ErrRunCacheMiss, input.Path)
	}
	if changed, inode := changeTimeAndInode(input.Path, information); changed != input.ChangedNanoseconds || inode != input.Inode {
		return fmt.Errorf("%w: %s was changed or replaced under its modification time", ErrRunCacheMiss, input.Path)
	}
	return nil
}

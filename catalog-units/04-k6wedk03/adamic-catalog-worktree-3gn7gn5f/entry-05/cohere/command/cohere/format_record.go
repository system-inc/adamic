package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/format/formatfiles"
	"github.com/system-inc/cohere/internal/gitignore"
	"github.com/system-inc/cohere/internal/release/packaging"
	"github.com/system-inc/cohere/internal/types/program"
)

// The format record: which bytes cohere has already seen formatted, so a run formats what changed
// since cohere last looked rather than what git says changed since the last commit.
//
// The scope used to be git's answer: working tree, staged and untracked, asked of every submodule in
// turn. That was a dozen subprocesses on ahra and the largest single cost outside the phases, and it
// answered a neighbouring question. A file committed unformatted was never in it, so it was never
// formatted and never reported, and a file edited and committed between two runs was never formatted
// either. The record answers the question the format phase actually has: are these bytes already known
// to be formatted?
//
// # What an entry proves
//
// An entry is the hash of a text the formatter returned unchanged, which is what formatTransform hands
// back: it formats until a pass changes nothing. So a file whose bytes hash to an entry is a fixed point
// of the formatter under the options it was recorded with, and formatting it again would change
// nothing. That holds whether or not the text was ever written. A `--no-fix` run over an unformatted
// file records the formatted text, the disk still holds the old bytes, the hashes differ, and the file
// stays in scope and keeps being reported. Nothing about the write has to be trusted.
//
// # What makes the record say nothing
//
// The record is a section of the project's cache table (see run_cache.go), keyed by the formatter's
// identity: a hash of the printers and everything they import, the pass loop, and the toolchain,
// stamped at build time. A cohere whose formatter changed starts with an empty record and formats every
// file once; one that changed only a lint rule keeps it. Each entry carries the options its file formats with, so a config edit sends exactly the
// files it reaches back through the formatter. Any doubt, an unreadable record, a version it does not
// know, is an empty record: the cost is one run that formats everything, never a file skipped on a
// guess.
//
// # Reading it cheaply
//
// An entry also holds the size and modification time its file had when its bytes were last read and
// matched, so an unchanged file is a stat rather than a read. That is the same trade the run cache
// makes, and it is stated in the same words there (see program.RunCache): a file whose bytes change
// while its size and nanosecond modification time do not is invisible to it.

type formatRecord struct {
	// root is the project root, whose cache table holds the record.
	root string

	// key is the formatter's identity (see formatRecordKey), empty when it could not be determined, in
	// which case nothing is read or written and every file is in scope. A section under another key
	// says nothing about this formatter.
	key string

	// absent says why no earlier check is on record, and is empty when one was read.
	absent string

	mutex   sync.Mutex
	entries map[string]program.FormatEntry

	// changed says an entry was added, replaced or given a signature since the record was read, so a
	// caller that writes only what moved (see saveChanged) knows whether there is anything to write.
	changed bool
}

// loadFormatRecord reads the record from the root's cache table. It never fails: a record that cannot
// be read is an empty one, and absent says why.
func loadFormatRecord(root string) *formatRecord {
	record := &formatRecord{root: root, entries: map[string]program.FormatEntry{}}
	key, err := formatRecordKey(root)
	if err != nil {
		record.absent = fmt.Sprintf("cohere could not identify its own binary (%v)", firstLine(err.Error()))
		return record
	}
	record.key = key

	// A table this cohere would discard, unreadable or written by another build, reads as no section.
	section := readFormatSection(root)
	switch {
	case section == nil:
		record.absent = "no earlier check is on record"
	case section.Key != key:
		record.absent = "the formatter changed since the last check, so its record says nothing about this one"
	case section.Entries != nil:
		record.entries = section.Entries
	}
	return record
}

// formatRecordOff is a record that holds nothing and keeps nothing, for a run that reads and writes no
// cache: every file is in scope, absent says why, and save writes nothing because there is no key.
func formatRecordOff(reason string) *formatRecord {
	return &formatRecord{absent: reason, entries: map[string]program.FormatEntry{}}
}

// formatRecordKey is what the record's section must have been written under to be read: the formatter's
// identity, stamped by the launcher at build time (see dispatch.FormatterIdentity), so a cohere commit
// that leaves the formatter alone leaves the record valid.
//
// The cache table no longer guards this section against a cohere commit (it keeps the section across one
// for this key to decide), so the key carries the whole claim. A build nobody stamped, such as a `go
// build` or a test binary, cannot say what its formatter is, so its key is the binary itself, which no
// other build shares.
func formatRecordKey(root string) (string, error) {
	if identity := release.Current().FormatterIdentity; identity != "" {
		return "formatter " + identity, nil
	}
	return program.RunCacheKey(nil, program.NewPathAnchor(root).Resolved(), "format-record")
}

// unformatted returns the files whose current bytes are not on record as formatted under the options
// they format with now, sorted.
func (record *formatRecord) unformatted(files []string, optionsFor func(fileName string) (string, error)) []string {
	confirmed := make([]bool, len(files))
	if record.key != "" {
		next := make(chan int, 256)
		var workers sync.WaitGroup
		for range min(runtime.NumCPU(), 8) {
			workers.Add(1)
			go func() {
				defer workers.Done()
				for index := range next {
					confirmed[index] = record.formatted(files[index], optionsFor)
				}
			}()
		}
		for index := range files {
			next <- index
		}
		close(next)
		workers.Wait()
	}

	unformatted := []string{}
	for index, fileName := range files {
		if !confirmed[index] {
			unformatted = append(unformatted, fileName)
		}
	}
	sort.Strings(unformatted)
	return unformatted
}

// formatted reports whether a file's bytes are on record as formatted under its current options.
func (record *formatRecord) formatted(fileName string, optionsFor func(fileName string) (string, error)) bool {
	record.mutex.Lock()
	entry, present := record.entries[fileName]
	record.mutex.Unlock()
	if !present {
		return false
	}
	options, err := optionsFor(fileName)
	if err != nil || options != entry.Options {
		return false
	}

	// Stat before reading. A write after the stat moves the modification time past the one about to be
	// recorded, so the next run sees a mismatch and reads again; the other order could record a
	// signature taken after bytes it never read.
	information, err := os.Stat(fileName)
	if err != nil || !information.Mode().IsRegular() {
		return false
	}
	if entry.ModifiedNanoseconds != 0 &&
		information.Size() == entry.Size && information.ModTime().UnixNano() == entry.ModifiedNanoseconds {
		return true
	}

	contents, err := os.ReadFile(fileName)
	if err != nil || formatRecordSum(contents) != entry.Sum {
		return false
	}

	// A file written within the last second is not given a signature, so a second write inside the
	// same tick of a coarse clock cannot hide behind it. It is read again next run instead.
	if time.Since(information.ModTime()) > time.Second {
		entry.Size = information.Size()
		entry.ModifiedNanoseconds = information.ModTime().UnixNano()
		record.mutex.Lock()
		record.entries[fileName] = entry
		record.changed = true
		record.mutex.Unlock()
	}
	return true
}

// vouches reports whether text is the bytes the record holds as formatted for a file, under the options
// it formats with now. Unlike formatted, it is asked about text in hand, such as a fixer's output, rather
// than the file on disk.
func (record *formatRecord) vouches(fileName string, text string, optionsFor func(fileName string) (string, error)) bool {
	record.mutex.Lock()
	entry, present := record.entries[fileName]
	record.mutex.Unlock()
	if !present {
		return false
	}
	options, err := optionsFor(fileName)
	return err == nil && options == entry.Options && formatRecordSum([]byte(text)) == entry.Sum
}

// observe wraps a format transform so every text it returns is recorded as formatted. A nil record or
// transform is handed back unchanged.
func (record *formatRecord) observe(inner edit.Transform, optionsFor func(fileName string) (string, error)) edit.Transform {
	if record == nil || inner == nil || record.key == "" {
		return inner
	}
	return func(fileName string, text string, parsed *ast.SourceFile) (string, error) {
		formatted, err := inner(fileName, text, parsed)
		if err != nil {
			return formatted, err
		}
		if options, optionsError := optionsFor(fileName); optionsError == nil {
			entry := program.FormatEntry{Sum: formatRecordSum([]byte(formatted)), Options: options}
			record.mutex.Lock()
			if previous, present := record.entries[fileName]; !present || previous.Sum != entry.Sum || previous.Options != entry.Options {
				record.entries[fileName] = entry
				record.changed = true
			}
			record.mutex.Unlock()
		}
		return formatted, nil
	}
}

// save writes the record into the root's cache table, leaving the table's other sections as they are.
// universe, when not nil, is every file a default-scope run could have formatted, and entries for
// anything outside it are dropped so the record does not keep files that were deleted or are now
// ignored. A run with a narrower scope passes nil and drops nothing.
func (record *formatRecord) save(universe []string) error {
	if record == nil || record.key == "" {
		return nil
	}
	record.mutex.Lock()
	entries := record.entries
	if universe != nil {
		inUniverse := make(map[string]struct{}, len(universe))
		for _, fileName := range universe {
			inUniverse[fileName] = struct{}{}
		}
		kept := make(map[string]program.FormatEntry, len(entries))
		for fileName, entry := range entries {
			if _, present := inUniverse[fileName]; present {
				kept[fileName] = entry
			}
		}
		entries = kept
	}
	section := &program.FormatSection{Key: record.key, Entries: entries}
	record.mutex.Unlock()
	return writeFormatSection(record.root, section)
}

// saveChanged is save for a reader that should leave the table alone when it learned nothing: it writes
// only when an entry moved or the universe drops one. The nested drift check reads repositories it never
// writes, so on an unchanged library it should not rewrite the library's table every run either.
func (record *formatRecord) saveChanged(universe []string) error {
	if record == nil || record.key == "" {
		return nil
	}
	record.mutex.Lock()
	moved := record.changed
	if !moved && universe != nil {
		inUniverse := make(map[string]struct{}, len(universe))
		for _, fileName := range universe {
			inUniverse[fileName] = struct{}{}
		}
		for fileName := range record.entries {
			if _, present := inUniverse[fileName]; !present {
				moved = true
				break
			}
		}
	}
	record.mutex.Unlock()
	if !moved {
		return nil
	}
	return record.save(universe)
}

func formatRecordSum(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}

// formatUniverse is every file the default format scope is drawn from: what the formatter handles in
// the repository the run writes, and nowhere else.
//
// It stops at the repository's own boundary. Writes stay inside one repository (@system_cohere,
// 2026-10-03): a submodule is formatted by a run inside it, which leaves the commit in the submodule,
// and a run here reads the submodules it declares and reports their drift instead of writing it (see
// checkNestedRepositories). The universe used to descend into every declared submodule, so a default
// run in ahra rewrote Structure and Nexus whenever their files were not on record, and the change
// landed in repositories nobody running here was committing to.
type formatUniverse struct {
	// root is the repository's own walk; its nested repositories are named and not walked.
	root formatfiles.Enumeration

	files []string
}

func enumerateFormatUniverse(engine formatEngine, root string) (formatUniverse, error) {
	rootEnumeration, err := enumerateFormatTree(engine, root)
	if err != nil {
		return formatUniverse{}, err
	}
	return formatUniverse{root: rootEnumeration, files: append([]string(nil), rootEnumeration.Files...)}, nil
}

// formatWalkAhead is the format walk begun before the graph build, for the fix phase to take.
type formatWalkAhead struct {
	engine      formatEngine
	root        string
	done        chan struct{}
	enumeration formatfiles.Enumeration
	err         error
}

// aheadFormatWalk is the walk startFormatWalkAhead began, nil once taken or when none was.
var aheadFormatWalk *formatWalkAhead

// startFormatWalkAhead begins engine's walk of root on its own goroutine and returns at once. The fix phase
// takes the result through enumerateFormatTree. It is the same call the fix phase would make, on the same
// engine and root, only earlier, so the enumeration is the one it would have made: the tree is not written
// between the two, since nothing writes before the fix phase does.
//
// listing, when not nil, is directory listings discovery already read, and an engine that can walk from them
// does (#g3046x5): the walk then finishes in tens of milliseconds instead of about 370ms on a cold ahra run,
// and the early format pass that waits on it starts that much sooner. The listings are a moment older than a
// walk of the disk would be, which the run cache's clock already covers, since it starts before discovery
// lists anything (see startRunCacheClock).
func startFormatWalkAhead(engine formatEngine, root string, listing gitignore.Listing) {
	ahead := &formatWalkAhead{engine: engine, root: root, done: make(chan struct{})}
	go func() {
		defer close(ahead.done)
		if listed, ok := engine.(listedEnumerator); ok && listing != nil {
			ahead.enumeration, ahead.err = listed.EnumerateListed(root, listing)
		} else {
			ahead.enumeration, ahead.err = engine.Enumerate(root)
		}
		ahead.enumeration = withoutYieldedDirectories(ahead.enumeration)
	}()
	aheadFormatWalk = ahead
}

// listedEnumerator is an engine whose walk can read listings already read (native.Resolving).
type listedEnumerator interface {
	EnumerateListed(root string, listing gitignore.Listing) (formatfiles.Enumeration, error)
}

// enumerateFormatTree is engine.Enumerate(root), taken from the walk begun ahead when that walk was of the same
// engine and root. A walk begun ahead is taken once; any other call walks the tree itself.
func enumerateFormatTree(engine formatEngine, root string) (formatfiles.Enumeration, error) {
	if ahead := aheadFormatWalk; ahead != nil && ahead.engine == engine && ahead.root == root {
		aheadFormatWalk = nil
		<-ahead.done
		return ahead.enumeration, ahead.err
	}
	// Less what the projects below this one format, when it is one project of several. See ownership.go.
	enumeration, err := engine.Enumerate(root)
	return withoutYieldedDirectories(enumeration), err
}

// declaredSubmodules reads the submodule paths a repository's `.gitmodules` declares. A repository with
// no `.gitmodules` declares none; one that cannot be read is an error, because reading it as empty would
// silently drop every submodule from the scope.
//
// Only the `path` key is read. The file is git's config format, and a value can be quoted; anything
// stranger than that names a path no read would find, which costs a submodule its drift check.
func declaredSubmodules(directory string) (map[string]struct{}, error) {
	declared := map[string]struct{}{}
	file, err := os.Open(filepath.Join(directory, ".gitmodules"))
	if errors.Is(err, os.ErrNotExist) {
		return declared, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading the submodules %s declares: %w", directory, err)
	}
	defer file.Close()
	lines := bufio.NewScanner(file)
	for lines.Scan() {
		key, value, found := strings.Cut(strings.TrimSpace(lines.Text()), "=")
		if !found || strings.TrimSpace(key) != "path" {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		declared[strings.TrimSuffix(value, "/")] = struct{}{}
	}
	if err := lines.Err(); err != nil {
		return nil, fmt.Errorf("reading the submodules %s declares: %w", directory, err)
	}
	return declared, nil
}

// describe says what the universe walked, in the words describeEnumeration uses for one walk.
func (universe formatUniverse) describe() string {
	return describeEnumeration(universe.root, len(universe.root.Files))
}

// unformattedScope is the default format scope: every file in the universe whose bytes are not on record
// as formatted. It returns the universe's files too, so the record can drop entries outside it.
func unformattedScope(engine formatEngine, record *formatRecord, root string) (formatScope, []string) {
	universe, err := enumerateFormatUniverse(engine, root)
	if err != nil {
		// A failed walk withholds formatting and says why, rather than falling back to a universe that
		// would format the wrong set. Fixing still runs. No universe means nothing is pruned, and nothing
		// says what the walk read, so the run is not replayed.
		declineRunCache("the format walk failed")
		return formatScope{
			index:       map[string]struct{}{},
			Description: fmt.Sprintf("nothing (could not enumerate the tree: %v)", err),
			failure:     err,
		}, nil
	}

	declareFormatWalk(universe.root)
	return unformattedScopeOf(engine, record, universe), universe.files
}

// unformattedScopeOf is unformattedScope's answer for a universe already walked, with no side effect on the run
// cache, so it can be computed off the main goroutine (startEarlyFormat) and declared there afterwards.
func unformattedScopeOf(engine formatEngine, record *formatRecord, universe formatUniverse) formatScope {
	unformatted := record.unformatted(universe.files, engine.OptionsFingerprint)
	index := make(map[string]struct{}, len(unformatted))
	for _, fileName := range unformatted {
		index[fileName] = struct{}{}
	}

	description := fmt.Sprintf("%d of %d files not on record as formatted at their current bytes",
		len(unformatted), len(universe.files))
	if record.absent != "" {
		description = fmt.Sprintf("all %d files, because %s", len(universe.files), record.absent)
	}
	return formatScope{
		FileNames:   unformatted,
		index:       index,
		Description: description + " · " + universe.describe(),
		recorded: func(fileName string, text string) bool {
			return record.vouches(fileName, text, engine.OptionsFingerprint)
		},
		adamic:      universe.root.Adamic,
		adamicJoins: true,
		adamicUnrecorded: func(files []string) []string {
			return record.unformatted(files, engine.OptionsFingerprint)
		},
	}
}

// optionsFingerprintOf is a formatter's options fingerprint, or, with no formatter, a function that
// refuses: there are no options to fingerprint, and nothing is recorded without one.
func optionsFingerprintOf(engine formatEngine) func(fileName string) (string, error) {
	if engine == nil {
		return func(string) (string, error) { return "", errors.New("no formatter was configured") }
	}
	return engine.OptionsFingerprint
}

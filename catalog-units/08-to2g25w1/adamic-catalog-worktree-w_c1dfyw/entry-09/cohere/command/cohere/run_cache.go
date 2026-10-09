package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/system-inc/cohere/internal/format/formatfiles"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/gitignore"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/release/packaging"
	"github.com/system-inc/cohere/internal/types/program"
)

// The run cache's command half: decide whether this run can be replayed, replay it if so, and
// otherwise record it on the way out.
//
// The package half (internal/types/program/run_cache.go) proves inputs unchanged. This half decides
// which runs are eligible and which exits are verdicts, and it is deliberately narrow on both.
//
// # Which runs
//
// A bare `cohere`, `cohere --no-fix`, and `cohere --no-fix --format`, and nothing else. Those are the runs
// people repeat, and every other flag widens the set of things the output could depend on. The key covers
// arguments anyway, so admitting more later is a change to one condition rather than to the proof, plus
// declaring whatever the new run reads that the graph build does not.
//
// # Which exits
//
// A run is recorded only when it ends in a verdict about the tree: clean, findings, or the type bail.
// A run that fails as a tool (an error out of run) is never recorded, since replaying it would replay a
// failure of the tool as if it were a fact about the code. And a run is recorded only if it declared
// its inputs, so any path that never reaches the declaration (Swift, listings, rename, stdin) cannot be
// recorded by construction rather than by a list someone has to keep complete.
//
// # What a replay prints
//
// The recorded report, framed. That report carries the recorded run's durations, and printed bare
// they would claim this run took seconds and built a graph it did not build. So a replay says when the
// report is from before it, and what this run actually cost after it.

// activeRunCache is the recording in progress, nil when this run is not being recorded.
var activeRunCache *runCacheSession

type runCacheSession struct {
	// directory is where this root's cache table lives, invocation the name this run's record goes under
	// in it, and table what was read from it when the run began: this invocation's run and the sections a
	// walk reuses.
	directory  string
	invocation string
	table      *program.CacheTable

	key      string
	recorder *program.InputRecorder

	// anchor is the project root, which the table names paths relative to and the recording names its
	// output's root by, so every spelling of the root is one cache (#547dhjz). See program.PathAnchor.
	anchor program.PathAnchor

	// keyParts is key's parts, recorded with the run so a later one under another key can say which moved.
	keyParts program.RunCacheKeyParts

	// missed is why this run was not replayed whole, and findingsMissed why its files' findings were not, when
	// the findings section as a whole was missing or under another key. Each is said in --verbose and --json, so
	// a run that could have replayed and did not explains itself (#547dhjz).
	missed         string
	findingsMissed string

	// findingsKeyParts is the findings key's parts, recorded with the section. See keyParts.
	findingsKeyParts program.RunCacheKeyParts

	// readSince is when the run began reading its inputs. An input changed after it keeps the run from
	// being recorded.
	readSince time.Time

	// declared is set once the build succeeds and the inputs are known. A session that never declares
	// is never recorded.
	declared   bool
	extraFiles []string

	// extraDirectories is every directory a format walk listed, so a file added, removed or renamed in one
	// moves an input. See declareFormatWalk.
	extraDirectories []string

	// extraAbsent is every file a format walk looked for and did not find, an ignore file a directory
	// does not have, so one appearing moves an input. See declareFormatWalk.
	extraAbsent []string

	// declined is the reason this run must not be recorded, empty while it may be.
	declined string

	// findings is the findings cache this run consults and records, the run cache's second layer. Nil
	// when the graph never reached a walk.
	findings *program.FindingsReuse

	// shapes is every project file's shape this run, recorded so the next computes only what changed. Nil
	// when no rule is keyed on shapes, since then nothing reads them and computing them would be waste.
	shapes map[string]program.SignatureEntry
	// shapesGraph is the graph shapes describe, the one rebuilt after a fix rewrite when there was one, kept so
	// the shapes still keyed on content can graduate after the verdict. See graduateSignatures.
	shapesGraph *program.Graph

	// types is the types phase's section, what the table held and what this run records. Nil when the
	// graph never reached a walk.
	types *program.TypeDiagnosticsReuse

	// summary is the run's summary, encoded, which a replay renders its footer from. See recordRunSummary.
	summary []byte

	stdout, stderr teeStream
}

// recordRunSummary keeps the run's summary for the record, when this run is being recorded.
func recordRunSummary(summary runSummary) {
	if activeRunCache == nil {
		return
	}
	if encoded, err := json.Marshal(summary); err == nil {
		activeRunCache.summary = encoded
	}
}

// teeStream copies a standard stream to its real destination and to a buffer.
type teeStream struct {
	real   *os.File
	writer *os.File
	buffer bytes.Buffer
	done   chan struct{}
}

func (t *teeStream) start(target **os.File) error {
	reader, writer, err := os.Pipe()
	if err != nil {
		return err
	}
	t.real = *target
	t.writer = writer
	t.done = make(chan struct{})
	go func() {
		defer close(t.done)
		defer reader.Close()
		// Line by line, so a line tagged as describing this invocation reaches the terminal untagged
		// and never reaches the replay. The tag travels in the stream itself because the tee is a pipe:
		// noting the buffer's length at the moment of printing would read it before the pipe delivered.
		lines := bufio.NewReader(reader)
		for {
			line, err := lines.ReadBytes('\n')
			if len(line) > 0 {
				if rest, tagged := bytes.CutPrefix(line, invocationTag); tagged {
					t.real.Write(rest)
				} else if rest, tagged := bytes.CutPrefix(line, provenanceTag); tagged {
					// Kept tagged in the recording, so a replay knows to say where the line came from.
					t.real.Write(rest)
					t.buffer.Write(line)
				} else {
					t.real.Write(line)
					t.buffer.Write(line)
				}
			}
			if err != nil {
				return
			}
		}
	}()
	*target = writer
	return nil
}

// stop restores the stream and waits for everything written so far to reach both destinations.
func (t *teeStream) stop(target **os.File) {
	if t.writer == nil {
		return
	}
	*target = t.real
	t.writer.Close()
	<-t.done
	t.writer = nil
}

// runCacheEligible reports whether this invocation may use the run cache at all.
//
// A bare run and `--no-fix`, each with `--format` or `--no-format` or neither, in any order. `--no-fix`
// runs every phase and writes nothing, so it is a pure report, which suits a cache better than the bare
// run does; it is what a real-tree measurement uses, so the probe asks to write nothing as well as being
// sandboxed. Both format by default (#b1sjy7b), so `--format` names what they already do, and their walk
// declares what it read (see declareFormatWalk); `--no-format` leaves formatting out. A bare run that
// rewrites a file declines its own record (see declineRunCache). The key covers the arguments, so none of
// them replays another.
//
// `--verbose`, `--json` and `--phases` choose how the run is printed and change nothing it computes, so
// each of them may join any of those runs. Each is still its own record, because each prints a different
// body: the key covers the arguments, and whether the output is colored.
//
// `--adamic-readiness` measures what the run without it does not, so it is its own record too, and a run
// with it never replays one without it, nor one without it a run with it (#9tgm3dq).
//
// `--no-cache` is refused by name rather than left to the argument shape. Most flags make a run
// ineligible, but the promise that flag makes, nothing read and nothing written, should not rest on which
// arguments happen to be admitted next.
func runCacheEligible() bool {
	if cacheOff {
		return false
	}
	seen := map[string]bool{}
	for _, argument := range os.Args[1:] {
		switch argument {
		case "--no-fix", "--format", "--no-format", "--verbose", "--json", "--phases", "--adamic-readiness":
		default:
			return false
		}
		if seen[argument] {
			return false
		}
		seen[argument] = true
	}
	// Refused by the command line before anything runs, and never a run to record.
	return !(seen["--format"] && seen["--no-format"]) && !(seen["--verbose"] && seen["--json"])
}

// beginRunCache replays a recorded run and exits if every input is unchanged, and otherwise starts
// recording this one. It returns the recorder the graph build should report its inputs to, or nil
// when this run is not eligible.
func beginRunCache(location projectLocation) *program.InputRecorder {
	if !runCacheEligible() {
		return nil
	}
	// No format scope is in the key. Without a formatter (`--no-format`) the scope is a constant. With one,
	// the scope is drawn from a walk that declares every file and directory it read, so a change that would
	// move the scope moves an input instead (see declareFormatWalk). A run that rewrites a file declines its
	// record. The printers need no fact of their own: the key covers the running binary, so any other build
	// of the formatter misses. Whether the output is colored is one, since a body recorded for a terminal
	// carries escape codes that a replay into a pipe must not print.
	// The root resolved, and the files below it from it, so a run through another spelling of the same
	// directory has the same key (#547dhjz).
	anchor := program.NewPathAnchor(location.Root)
	keyParts, err := program.RunCacheKeyPartsOf(os.Args[1:], anchor.Resolved(),
		"root="+anchor.Resolved(),
		"tsconfig="+anchor.Stable(location.ConfigFileName),
		"lint-config="+anchor.Stable(location.LintConfigFileName),
		fmt.Sprintf("color=%t", activeOutput.Style.color),
		// What the run leaves to nearer projects changes what it reports, and it arrives by a variable the
		// arguments do not show. See ownership.go.
		"yield="+activeProjectYieldFingerprint(),
	)
	if err != nil {
		return nil
	}
	key := keyParts.Key()

	prepareCacheDirectory(location.Root)
	directory := cacheDirectory(location.Root)
	printPreviousNotes(directory)
	// The run before this one may still be writing the table after answering its caller, and reading
	// before its renames would miss what it recorded, or see its run beside the findings before them. So the
	// lock is held shared across every file read here. See table_lock.go.
	release, _ := holdTableReadLock(directory, tableWriterWait)
	identity := cacheTableIdentity()
	// This invocation's run first and alone: a replay needs nothing else, and the run is a sixth of the
	// table's bytes where the whole table was all of them (#45ekc65).
	invocation := program.CacheTableInvocation(os.Args[1:])
	table, err := program.ReadCacheTable(directory, identity, program.CacheTableSections{Runs: []string{invocation}}, anchor)
	stored := table.Runs[invocation]
	// Every input's stat is noted as it is checked, so a run that misses hands them to the content pack rather
	// than statting each file again (#kdee854). See program.StatSnapshot.
	runCacheCheckStats = program.NewStatSnapshot()
	checked := stored.CheckNoting(key, runCacheCheckStats)
	if checked == nil {
		release()
		replayRunCache(stored, anchor)
	}
	sections, sectionsError := program.ReadCacheTable(directory, identity,
		program.CacheTableSections{Findings: true, Signatures: true, Types: true}, anchor)
	release()
	table.Findings, table.Signatures, table.Types = sections.Findings, sections.Signatures, sections.Types
	for _, readError := range []error{err, sectionsError} {
		if errors.Is(readError, program.ErrCacheTableUnreadable) || errors.Is(readError, program.ErrCacheTablePartlyKept) {
			// Said once, before the recording starts, so it reaches the terminal and never a replay. A table
			// thrown away on every run would otherwise look like a cache that is merely cold.
			fmt.Fprintf(accountOutput(os.Stderr), "note: %v; this run starts cold and writes what it drops anew\n", readError)
		}
	}

	session := &runCacheSession{
		directory:  directory,
		invocation: invocation,
		table:      table,
		key:        key,
		keyParts:   keyParts,
		missed:     runCacheMiss(stored, checked, keyParts, err),
		anchor:     anchor,
		recorder:   program.NewInputRecorder(),
		// See startRunCacheClock: before discovery reads anything, after the cache directory exists.
		readSince: runCacheClock,
	}
	if err := session.stdout.start(&os.Stdout); err != nil {
		return nil
	}
	if err := session.stderr.start(&os.Stderr); err != nil {
		session.stdout.stop(&os.Stdout)
		return nil
	}
	activeRunCache = session
	return session.recorder
}

// cacheDirectory is where cohere keeps what it caches for one project, the cache table's files among it:
// `<root>/.cache/cohere/`.
//
// It lived in the user cache under a hash of the root path, which nothing ever reclaimed. A moved or
// deleted project left its table behind for good, and 38 of them had piled up by 2026-10-02. In the
// project it is one folder per project that deleting is always a correct answer for, and a worktree has
// its own because it is its own root. Every eligible invocation shares it; their recorded runs sit side
// by side, each in its own file, and their findings are the same findings.
func cacheDirectory(root string) string {
	return filepath.Join(root, ".cache", "cohere")
}

// prepareCacheDirectory creates the project's cache directory before anything is read or recorded, and
// warns once when the project's .gitignore does not ignore it.
//
// Created first because the run cache records the project root's directory as an input, at modification
// time. Creating `.cache` moves the root's time, and doing it after the inputs were recorded would make
// the first replay of every project a miss that looks merely cold. Writes inside `.cache/cohere` move only
// that directory's time, which no tsconfig enumerates, since a wildcard never matches a dot directory.
//
// The warning is a warning rather than a refusal: the cache is still the project's to keep, and the line
// to add is named so it is one edit away. What it must never be is a silent write into tracked space.
var prepareCacheDirectoryOnce sync.Once

// runCacheClock is when the run began reading its inputs, the time a recording's inputs must not have changed
// since. Set by startRunCacheClock.
var runCacheClock time.Time

// startRunCacheClock starts the run cache's clock: after the cache directory exists, since creating it moves the
// project root's time, and before discovery lists the tree. The clock was taken at the session's start, in
// beginRunCache, until discovery's listings came to serve the build's enumeration (#bjv0tg4); discovery runs
// before the session, so a file created between its listing and the session's start was missing from the
// enumeration while its directory's time already counted it, and the next run replayed over it
// (TestAFileCreatedAfterDiscoveryIsNeverReplayedOver). An earlier clock costs only a run whose inputs change in
// that moment, which is then not recorded.
func startRunCacheClock(location projectLocation, locateError error) {
	if locateError == nil && runCacheEligible() {
		prepareCacheDirectory(location.Root)
	}
	// Any run that keeps a cache may read the libraries' records, eligible for replay or not.
	if locateError == nil && !cacheOff {
		prepareNestedCacheDirectories(writeRepositoryRoot(location.ArgumentBase, location.Root))
	}
	runCacheClock = time.Now()
}

func prepareCacheDirectory(root string) {
	prepareCacheDirectoryOnce.Do(func() {
		if err := os.MkdirAll(cacheDirectory(root), 0o755); err != nil {
			fmt.Fprintf(accountOutput(os.Stderr), "note: the cache directory %s could not be created: %v\n", cacheDirectory(root), firstLine(err.Error()))
			return
		}
		ignoreFile := filepath.Join(root, gitignore.IgnoreFileName)
		matcher, err := gitignore.New(root)
		if err != nil {
			return
		}
		if ignored, _, err := matcher.IgnoredPath(".cache/cohere/findings.gob", false); err == nil && !ignored {
			fmt.Fprintf(accountOutput(os.Stderr), "note: %s does not ignore .cache/, so cohere's cache in %s would be tracked: add the line `.cache/` to it\n",
				ignoreFile, cacheDirectory(root))
		}
	})
}

// readFormatSection is the format record's section of this root's table, nil when there is none or the
// table was discarded. The caller decides whether its key still holds.
func readFormatSection(root string) *program.FormatSection {
	if cacheOff {
		return nil
	}
	release, _ := holdTableReadLock(cacheDirectory(root), tableWriterWait)
	table, _ := program.ReadCacheTable(cacheDirectory(root), cacheTableIdentity(), program.CacheTableSections{Formatted: true},
		program.NewPathAnchor(root))
	release()
	return table.Formatted
}

// writeFormatSection replaces the format record's file and touches no other. The record is saved where the
// format phase ends, mid-run, for any run that formats, writing or not; its own file makes that a write of
// the record alone, about a megabyte on ahra, where it used to be a read and a write of the whole table
// (#45ekc65).
func writeFormatSection(root string, section *program.FormatSection) error {
	if cacheOff {
		return nil
	}
	prepareCacheDirectory(root)
	directory := cacheDirectory(root)
	release := holdTableLock(directory)
	defer release()
	return program.WriteCacheTable(directory, &program.CacheTable{Formatted: section}, cacheTableIdentity(),
		program.CacheTableSections{Formatted: true}, program.NewPathAnchor(root))
}

// dumpCacheTable is `--cache-dump`: what this project's table holds, read by the same build that would use
// it, so a table this binary would discard says so rather than printing as though it were in use.
func dumpCacheTable(location projectLocation) error {
	directory := cacheDirectory(location.Root)
	release, _ := holdTableReadLock(directory, tableWriterWait)
	table, err := program.ReadCacheTable(directory, cacheTableIdentity(), program.EveryCacheTableSection, program.NewPathAnchor(location.Root))
	release()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("no cache table for %s (looked in %s)\n", location.Root, directory)
			return nil
		}
		fmt.Printf("%v\n", err)
		if !errors.Is(err, program.ErrCacheTablePartlyKept) {
			return nil
		}
	}
	program.DumpCacheTable(os.Stdout, directory, table, cacheTableIdentity())
	return nil
}

// cacheTableIdentity is the build a table must have been written by to be read. Anything about the binary
// that can change what a cached answer means is in it, and the keys inside the table cover the rest.
func cacheTableIdentity() program.CacheTableIdentity {
	provenance := release.Current()
	return program.CacheTableIdentity{
		SelfCommit:     provenance.SelfCommit,
		CompilerCommit: provenance.CompilerCommit,
		GoToolchain:    provenance.GoToolchain,
		Platform:       provenance.Platform,
	}
}

// runCacheMiss says why a run was not replayed whole, empty when it was about to be: no run on record for these
// arguments, one under another key and which part moved, or the input that changed. A table thrown away is
// said first, since then nothing else was read.
func runCacheMiss(stored *program.RunCache, checked error, parts program.RunCacheKeyParts, readError error) string {
	if checked == nil {
		return ""
	}
	discarded := ""
	if errors.Is(readError, program.ErrCacheTableUnreadable) || errors.Is(readError, program.ErrCacheTablePartlyKept) {
		discarded = firstLine(readError.Error()) + "; "
	}
	switch {
	case stored == nil:
		return discarded + "no run is on record for these arguments"
	case stored.Key != parts.Key():
		return discarded + "the run on record was taken under another key: " + parts.Changed(stored.KeyParts)
	}
	return discarded + strings.TrimPrefix(checked.Error(), program.ErrRunCacheMiss.Error()+": ")
}

// replayRunCache prints a recorded run, framed so its durations cannot be read as this run's, and
// exits with its exit code. Its paths are printed as this run spells the root (see recordedOutput).
func replayRunCache(stored *program.RunCache, anchor program.PathAnchor) {
	recorded := time.Unix(0, stored.RecordedUnixNanoseconds).Format("15:04:05")
	account := accountOutput(os.Stdout)
	fmt.Fprintf(account,
		"cached: no input has changed since the run at %s, so graph, types and lint did not run; its verdict follows\n",
		recorded)
	os.Stdout.Write(replayLines(spelledOutput(stored.Output, anchor), recorded))
	os.Stderr.Write(spelledOutput(stored.Errors, anchor))
	fmt.Fprintf(account, "phases: replayed the run at %s · fix, types and lint did not run\n", recorded)
	fmt.Fprintf(account, "  this run: %s, after checking %d inputs\n", round(time.Since(processStart)), len(stored.Inputs))
	fmt.Fprintf(account, "  %s\n", activeMemoryPolicy.line())

	// The footer is this replay's, rendered from the recorded run's summary: its findings and its counts,
	// with this run's time, nothing checked fresh, every file answered by the cache and nothing rewritten.
	// The recorded run's own footer is not in the recording, since it would claim that run's time.
	var summary runSummary
	if len(stored.Summary) > 0 && json.Unmarshal(spelledOutput(stored.Summary, anchor), &summary) == nil {
		summary = replayedSummary(summary, time.Since(processStart))
		switch activeOutput.Mode {
		case outputVerbose:
			fmt.Fprintln(os.Stdout, footer(summary, activeOutput.Style, footerOptions{Phases: activeOutput.Phases, Verbose: true}))
		case outputJSON:
			writeJSONLine(os.Stdout, summaryAsJSON(summary))
		default:
			fmt.Fprintln(os.Stdout, footer(summary, activeOutput.Style, footerOptions{Phases: activeOutput.Phases}))
		}
	}
	exitProcess(stored.ExitCode)
}

// replayedSummary is a recorded run's summary as its replay says it: the same findings and gaps, this
// replay's time, no phase run, nothing checked fresh, every file in scope answered by the cache, and
// nothing rewritten or walked. The recorded run's phases that ran are dropped, since none ran now; the
// ones that did not are kept, since a phase the recorded run never reached is a gap in the verdict this
// replay repeats.
func replayedSummary(recorded runSummary, elapsed time.Duration) runSummary {
	replayed := recorded
	replayed.Total = elapsed
	replayed.Graph, replayed.Formatting = 0, 0
	replayed.Phases = nil
	for _, record := range recorded.Phases {
		if record.Outcome == outcomeSkipped || record.Outcome == outcomeNotReached {
			replayed.Phases = append(replayed.Phases, record)
		}
	}
	// Every file in scope was answered by the cache, so the count says so: --json read 0 here, the number a run
	// that replayed nothing reads, and a reader took a whole replay for a full check (#547dhjz).
	replayed.Cache = cacheUse{Replayed: true, FilesReplayed: recorded.FilesInScope}
	replayed.FilesChecked, replayed.FilesCached = 0, recorded.FilesInScope
	replayed.Changed = nil
	replayed.Nodes = 0
	return replayed
}

// declareRunCacheInputs marks the build as having succeeded, and adds inputs the command reads itself.
func declareRunCacheInputs(files ...string) {
	if activeRunCache == nil {
		return
	}
	activeRunCache.declared = true
	activeRunCache.extraFiles = append(activeRunCache.extraFiles, files...)
}

// declareLintConfigInputs adds the lint config and every file it extends to this run's inputs, each present or
// absent by what is on disk. Zero config is the default settings file missing (#e0tybvv): declared as a file the
// run read, recording stat'ed a file that was not there and refused the record, so a zero-config project never
// replayed. Its absence is the input, since creating the file changes what runs.
func declareLintConfigInputs(path string) {
	session := activeRunCache
	if session == nil {
		return
	}
	for _, source := range configuration.SourcesOnDisk(lintConfigSources(path)) {
		if _, err := os.Stat(source); err == nil {
			session.extraFiles = append(session.extraFiles, source)
		} else {
			session.extraAbsent = append(session.extraAbsent, source)
		}
	}
}

// declareFormatWalk adds what one format walk depends on to this run's inputs, so a run that checks
// formatting can be replayed (#13a63n3).
//
// What the walk's answer is a function of: the files it found, whose bytes decide what the formatter
// reports; every directory it listed, whose modification time moves when an entry is added, removed or
// renamed, which is also what catches a new .git, settings file or leftover Prettier config inside the
// tree; and the files it read or would read, each recorded when present, since its directory already
// covers its appearing. Those are the root's .gitmodules, the .prettierignore the walk refuses, and every
// CohereSettings.json a directory resolves its options from, with the files it extends.
//
// The git ignore files the walk read or looked for (Enumeration.IgnoreFiles: each entered directory's
// .gitignore and info/exclude) are recorded present or absent. A content edit to a nested .gitignore moves
// no directory's modification time, and info/exclude sits in .git/info, a directory the walk never lists.
// The formatter's own identity needs no input: the run's key covers the running binary (see beginRunCache).
func declareFormatWalk(enumeration formatfiles.Enumeration) {
	session := activeRunCache
	if session == nil {
		return
	}
	for _, ignoreFile := range enumeration.IgnoreFiles {
		if _, err := os.Stat(ignoreFile); err == nil {
			session.extraFiles = append(session.extraFiles, ignoreFile)
		} else {
			session.extraAbsent = append(session.extraAbsent, ignoreFile)
		}
	}
	read := []string{
		filepath.Join(enumeration.Root, ".gitmodules"),
		filepath.Join(enumeration.Root, ".prettierignore"),
	}
	for _, directory := range enumeration.Directories {
		read = append(read, filepath.Join(directory, formatoptions.SettingsFileName))
	}
	// The settings the root itself resolves to can sit above it, in a directory the walk never listed, and so
	// can the directories between: a settings file created in one of those would take over the options.
	if resolution, err := formatoptions.Resolve(enumeration.Root); err == nil && resolution.Source != "" {
		read = append(read, resolution.Source)
		for directory := filepath.Clean(enumeration.Root); directory != filepath.Dir(resolution.Source); directory = filepath.Dir(directory) {
			session.extraDirectories = append(session.extraDirectories, directory)
			if parent := filepath.Dir(directory); parent == directory {
				break
			}
		}
	}
	for _, path := range read {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if filepath.Base(path) == formatoptions.SettingsFileName {
			session.extraFiles = append(session.extraFiles, configuration.SourcesOnDisk(lintConfigSources(path))...)
			continue
		}
		session.extraFiles = append(session.extraFiles, path)
	}
	session.extraFiles = append(session.extraFiles, enumeration.Files...)
	session.extraDirectories = append(session.extraDirectories, enumeration.Directories...)
}

// declineRunCache records why this run must not be replayed. The first reason stands.
func declineRunCache(reason string) {
	if activeRunCache != nil && activeRunCache.declined == "" {
		activeRunCache.declined = reason
	}
}

// abandonRunCache stops recording without writing anything, for a run that failed as a tool. It
// restores the real streams first, so the error about to be printed reaches the terminal.
func abandonRunCache() {
	session := activeRunCache
	if session == nil {
		return
	}
	activeRunCache = nil
	session.stdout.stop(&os.Stdout)
	session.stderr.stop(&os.Stderr)
}

// finishRunCache ends a run whose exit is a verdict about the tree: it records the run if it may, then
// exits with the code. Every verdict exit goes through here, or the cache would record a run with the
// wrong exit code, and a failing tree replayed as exit 0 passes CI while printing a failure.
func finishRunCache(exitCode int) {
	session := activeRunCache
	if session != nil {
		activeRunCache = nil
		session.stdout.stop(&os.Stdout)
		session.stderr.stop(&os.Stderr)
		// The report is out, so the caller can have its answer while the run is recorded. See sendVerdict.
		// The table's lock is taken first and held until the table is in place, so a run the caller starts
		// next waits for this write rather than reading around it. See table_lock.go.
		release := holdTableLock(session.directory)
		answered := sendVerdict(exitCode)
		var recorded *program.RunCache
		if session.declared && session.declined == "" {
			recorded = session.record(exitCode)
		}
		if answered {
			graduateSignatures(session.directory, session.shapesGraph, session.shapes)
		}
		session.write(recorded)
		release()
	} else if record := pendingTypes; record != nil {
		// A run nothing records still leaves the types section for the next, after its caller has the answer.
		pendingTypes = nil
		release := holdTableLock(record.directory)
		if sendVerdict(exitCode) {
			record.computedShapes += graduateSignatures(record.directory, record.shapesGraph, record.shapes)
		}
		record.write()
		release()
	}
	exitProcess(exitCode)
}

// signatureGraduationBudget is how long a run spends giving content-keyed shapes real signatures after its
// verdict is out. The table's lock is held throughout, so a run started at once waits this long at most.
var signatureGraduationBudget = time.Second

// graduateSignatures gives the shapes still keyed on content real declaration signatures, for at most
// signatureGraduationBudget, and returns how many entries changed, the abandoned one included (#9knyr86).
//
// Only after the verdict reached the dispatcher, which is what lets the emit's cost go unseen: a run nobody
// answered early is still a caller's wait. A project with no build info used to keep every shape keyed on
// content, so a comment in one file re-ran its importers' shape-keyed rules and re-checked their types; on ahra,
// 3,974 of 3,976 files, and an edit run's types phase of about 250ms against about 50ms once they graduate.
func graduateSignatures(directory string, graph *program.Graph, shapes map[string]program.SignatureEntry) int {
	if graph == nil || shapes == nil {
		return 0
	}
	graduated, abandoned := graph.GraduateSignatures(shapes, time.Now().Add(signatureGraduationBudget), signatureGraduationBudget/2)
	if abandoned != "" {
		cacheNote(directory, fmt.Sprintf("%s's declaration emit ran past %s, so its shape stays keyed on content until its bytes change",
			abandoned, signatureGraduationBudget/2))
		return graduated + 1
	}
	return graduated
}

// record captures this run, or returns nil when it cannot.
func (session *runCacheSession) record(exitCode int) *program.RunCache {
	// A replay renders its footer from the summary, so a run that recorded none would replay a verdict with
	// no footer. Refused here, once, rather than resting on every verdict path remembering writeRunEnd.
	if session.summary == nil {
		session.note("a run that recorded no summary is not recorded: its replay would print a verdict with no footer")
		return nil
	}
	present, absent, probed := session.recorder.Inputs()
	files := append(present, session.extraFiles...)
	absent = append(absent, session.extraAbsent...)
	cache, err := program.RecordRunCache(session.key, files, session.extraDirectories, absent, probed,
		recordedOutput(session.stdout.buffer.Bytes(), session.anchor), exitCode, session.readSince)
	if err != nil {
		// Not recording is always safe. Said on stderr because a cache that silently never records is
		// a saving that quietly never appears.
		session.note(fmt.Sprintf("the run cache did not record this run: %v", firstLine(err.Error())))
		return nil
	}
	cache.Errors = recordedOutput(session.stderr.buffer.Bytes(), session.anchor)
	cache.KeyParts = session.keyParts
	cache.RecordedUnixNanoseconds = time.Now().UnixNano()
	cache.Summary = recordedOutput(session.summary, session.anchor)
	return cache
}

// write puts this run's record, findings, shapes and types into the table, each into its own file, once.
//
// Only this run's files are written: its own invocation's run, and the sections it produced. Another
// invocation's run is another file, so a bare run and `--no-fix` started together never erase each other's
// record, and nothing has to be read back first to keep it. Two runs writing the same section is a race the
// later one wins, which costs a miss; nothing here can make a stale entry match, since every entry carries
// its own proof.
//
// The findings are saved even when the run itself was declined. Their entries are keyed on each file's
// bytes, so a file the fix phase rewrote left an entry for bytes that no longer exist, which can never
// match.
func (session *runCacheSession) write(recorded *program.RunCache) {
	table := program.NewCacheTable()
	var sections program.CacheTableSections
	if recorded != nil {
		table.Runs[session.invocation] = recorded
		sections.Runs = []string{session.invocation}
	}
	if session.findings != nil {
		table.Findings = session.findings.Recorded()
		table.Findings.KeyParts = session.findingsKeyParts
		sections.Findings = true
	}
	if session.shapes != nil {
		table.Signatures = session.shapes
		sections.Signatures = true
	}
	if session.types != nil {
		if recorded := session.types.Recorded(); recorded != nil {
			table.Types = recorded
			sections.Types = true
		}
	}
	if err := program.WriteCacheTable(session.directory, table, cacheTableIdentity(), sections, session.anchor); err != nil {
		session.note(fmt.Sprintf("the cache table could not be written: %v", firstLine(err.Error())))
	}
}

// note says something about recording this run. Before the caller has its verdict it goes to stderr, as
// it always did; after, the caller has moved on and its terminal may hold someone else's prompt, so it is
// kept beside the table for the next run to say instead (printPreviousNotes).
func (session *runCacheSession) note(text string) {
	cacheNote(session.directory, text)
}

// cacheNote says something about the cache in directory: on stderr while the caller is still waiting, and to
// the next run once it has its verdict.
func cacheNote(directory string, text string) {
	if !verdictSent {
		fmt.Fprintf(accountOutput(os.Stderr), "note: %s\n", text)
		return
	}
	file, err := os.OpenFile(previousNotesPath(directory), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	fmt.Fprintln(file, text)
	file.Close()
}

// previousNotesPath is where a run that answered its caller early keeps what it had to say afterward.
func previousNotesPath(directory string) string {
	return filepath.Join(directory, "notes-after-return.txt")
}

// printPreviousNotes says, once, what the last run recorded after it had already returned, and clears it.
// A cache write that failed in the background otherwise fails silently forever, the saving quietly gone.
func printPreviousNotes(directory string) {
	path := previousNotesPath(directory)
	contents, err := os.ReadFile(path)
	if err != nil {
		return
	}
	os.Remove(path)
	for line := range strings.SplitSeq(strings.TrimSpace(string(contents)), "\n") {
		if line != "" {
			fmt.Fprintf(accountOutput(os.Stderr), "note: after the last run returned, %s\n", line)
		}
	}
}

func firstLine(text string) string {
	if index := strings.IndexByte(text, '\n'); index >= 0 {
		return text[:index]
	}
	return text
}

// invocationTag marks a line that is true of this invocation and false of any replay of it.
//
// A recorded report says "graph built in 436ms" and "types ran in 409ms". Replayed, those lines claim
// a graph was built that was not, and durations this run did not spend. So each line describing the
// invocation rather than the tree is written through invocationOutput, which tags it; the tee strips
// the tag before the terminal and keeps the line out of what is replayed. The tag is a record
// separator, which no line of cohere's output contains.
var invocationTag = []byte("\x1ecohere-invocation\x1e")

// invocationOutput returns where to write a line that describes this invocation rather than the tree.
// Without a recording it is out itself, so a run that is not recorded prints exactly what it always did.
func invocationOutput(out io.Writer) io.Writer {
	if activeRunCache == nil {
		return out
	}
	return &taggingWriter{out: out, tag: invocationTag, atLineStart: true}
}

// taggingWriter puts its tag at the start of every line it writes.
type taggingWriter struct {
	out         io.Writer
	tag         []byte
	atLineStart bool
}

func (w *taggingWriter) Write(data []byte) (int, error) {
	written := 0
	for len(data) > 0 {
		if w.atLineStart {
			if _, err := w.out.Write(w.tag); err != nil {
				return written, err
			}
			w.atLineStart = false
		}
		end := bytes.IndexByte(data, '\n')
		chunk := data
		if end >= 0 {
			chunk = data[:end+1]
			w.atLineStart = true
		}
		n, err := w.out.Write(chunk)
		written += n
		if err != nil {
			return written, err
		}
		data = data[len(chunk):]
	}
	return written, nil
}

// provenanceTag marks a line that is true of the tree but was produced by a phase, so a replay prints it
// with where it came from.
//
// "fix: 0 of 5 files rewritten, 5 not formatted" is true of an unchanged tree and its not-formatted count
// is actionable, so a replay keeps it. Printed bare it would claim the fix phase ran on a run where
// nothing ran, so it is replayed as "fix (from the cached run at 03:41): ...". Unlike invocationTag the
// line stays in the recording, tagged, so the replay can find it without matching on its text.
var provenanceTag = []byte("\x1ecohere-provenance\x1e")

// provenanceOutput returns where to write such a line. Without a recording it is out itself.
func provenanceOutput(out io.Writer) io.Writer {
	if activeRunCache == nil {
		return out
	}
	return &taggingWriter{out: out, tag: provenanceTag, atLineStart: true}
}

// replayLines is a recording made printable: each provenance-tagged line "label: rest" becomes
// "label (from the cached run at recorded): rest", and no tag survives.
func replayLines(output []byte, recorded string) []byte {
	var replayed bytes.Buffer
	for _, line := range bytes.SplitAfter(output, []byte("\n")) {
		rest, tagged := bytes.CutPrefix(line, provenanceTag)
		if !tagged {
			replayed.Write(line)
			continue
		}
		label, detail, found := bytes.Cut(rest, []byte(": "))
		if !found {
			// A line with no label still must not claim to be current, so the source goes in front.
			fmt.Fprintf(&replayed, "(from the cached run at %s) %s", recorded, rest)
			continue
		}
		fmt.Fprintf(&replayed, "%s (from the cached run at %s): %s", label, recorded, detail)
	}
	return replayed.Bytes()
}

// rootTag stands for the project root in a recorded run's output, so a replay prints its paths as the replaying
// run spells the root rather than as the recording run did (#547dhjz).
var rootTag = []byte("\x1ecohere-root\x1e")

// recordedOutput is output with the project root, by either the spelling this run was given or its resolved
// target, replaced by rootTag. Only a whole path component is replaced: a root of /work/app leaves
// /work/application alone.
func recordedOutput(output []byte, anchor program.PathAnchor) []byte {
	for _, root := range []string{anchor.Root(), anchor.Resolved()} {
		if root == "" {
			continue
		}
		output = replaceRoot(output, []byte(root), rootTag)
	}
	return output
}

// spelledOutput is a recording's output with rootTag replaced by the root as this run spells it.
func spelledOutput(output []byte, anchor program.PathAnchor) []byte {
	return bytes.ReplaceAll(output, rootTag, []byte(anchor.Root()))
}

// replaceRoot replaces each occurrence of root that is a whole path: one that starts where a path does, not
// after a separator or inside a name (/tmp/x inside /private/tmp/x is not a root), and ends where a path
// component does, at a separator, at the end, or at a byte no name would continue with, such as a space, a
// colon or a comma.
func replaceRoot(output []byte, root []byte, replacement []byte) []byte {
	var replaced bytes.Buffer
	for {
		position := bytes.Index(output, root)
		if position < 0 {
			replaced.Write(output)
			return replaced.Bytes()
		}
		end := position + len(root)
		if position > 0 && !endsEscape(output[:position]) &&
			(output[position-1] == '/' || output[position-1] == '\\' || continuesName(output[position-1])) ||
			end < len(output) && continuesName(output[end]) {
			replaced.Write(output[:end])
			output = output[end:]
			continue
		}
		replaced.Write(output[:position])
		replaced.Write(replacement)
		output = output[end:]
	}
}

// endsEscape reports whether text ends with a terminal's color sequence, ESC [ digits and semicolons, m, which a
// colored report prints right before a path.
func endsEscape(text []byte) bool {
	// A color sequence is a few bytes, so only the tail is searched: a report with thousands of paths and no
	// color would otherwise be scanned back to its start for each one.
	text = text[max(0, len(text)-16):]
	start := bytes.LastIndexByte(text, 0x1b)
	if start < 0 || len(text)-start < 3 || text[start+1] != '[' || text[len(text)-1] != 'm' {
		return false
	}
	for _, character := range text[start+2 : len(text)-1] {
		if character != ';' && (character < '0' || character > '9') {
			return false
		}
	}
	return true
}

// continuesName reports whether a byte after a root would make it part of a longer name.
func continuesName(next byte) bool {
	return next == '-' || next == '_' || next == '.' || next >= '0' && next <= '9' || next >= 'A' && next <= 'Z' ||
		next >= 'a' && next <= 'z' || next >= 0x80
}

// attachFindingsCache gives this run's graph the findings cache, the run cache's second layer: when
// the run as a whole cannot be replayed, files whose bytes are unchanged replay their cacheable rules'
// findings and coverage, and only their uncacheable rules are walked. See program.FindingsReuse.
//
// Only a run being recorded gets one, so it is exactly as narrow as the run cache.
func attachFindingsCache(graph *program.Graph, location projectLocation) {
	session := activeRunCache
	if session == nil || graph == nil {
		return
	}
	parts, err := findingsCacheKey(graph, location)
	if err != nil {
		// Without a key nothing can be proven unchanged, so nothing is replayed or recorded.
		return
	}
	key := sha256.Sum256([]byte(parts.Key()))
	session.findingsKeyParts = parts
	switch previous := session.table.Findings; {
	case previous == nil:
		session.findingsMissed = "no findings are on record"
	case previous.Key != key:
		session.findingsMissed = "the findings on record were taken under another key: " + parts.Changed(previous.KeyParts)
	}
	session.findings = program.NewFindingsReuse(key, session.table.Findings, session.anchor)

	// Shapes cost a declaration emit for each file whose bytes changed, so they are computed only when some
	// rule is keyed on them. They start from the table's, then from the compiler's own buildinfo, so only
	// what changed since either is computed.
	if anyRuleKeyedOnShapes(registry.All()) {
		shapes, _ := graph.Signatures(context.Background(), graph.SeedSignatures(session.table.Signatures))
		graph.Shapes = shapes
		session.shapes, session.shapesGraph = shapes, graph
	}
	graph.FindingsReuse = session.findings
	// An entry recorded by a run that did not measure readiness was never measured, so a run that does must
	// miss on it rather than replay it as clean (#drbrp8c).
	if graph.Readiness != nil {
		session.findings.MeasureReadiness()
	}
	if typesKey, err := typesCacheKey(graph, location); err == nil {
		session.types = program.NewTypeDiagnosticsReuse(session.table.Types, typesKey)
	}
}

// carryCaches gives the graph rebuilt after a fix rewrite what the graph it replaces was given: the findings
// cache, and the shapes, recomputed only for the files whose bytes the fixer changed (#891h54d).
//
// The rebuild used to leave both off. With no shapes the types section cannot replay, so a run whose fixer
// rewrote one file full-checked the whole program, 380 to 1,270ms on ahra against 42 to 72ms replayed; with
// no findings cache, lint walked every rule over every file. Shapes are keyed per file on its bytes
// (Graph.Signatures), so seeding from the replaced graph's recomputes exactly the rewritten files, and the
// findings cache is per file on bytes and fingerprints, so the one cache serves both graphs.
func carryCaches(replaced *program.Graph, rebuilt *program.Graph) {
	rebuilt.FindingsReuse = replaced.FindingsReuse
	if replaced.Shapes == nil {
		return
	}
	shapes, computed := rebuilt.Signatures(context.Background(), replaced.Shapes)
	rebuilt.Shapes = shapes
	if session := activeRunCache; session != nil && session.shapes != nil {
		session.shapes, session.shapesGraph = shapes, rebuilt
	}
	if record := pendingTypes; record != nil {
		record.shapes, record.shapesGraph = shapes, rebuilt
		record.computedShapes += computed
	}
}

// activeTypesReuse is the types section a check in this run reads and records, nil when the cache is off
// or the run never reached a check.
func activeTypesReuse() *program.TypeDiagnosticsReuse {
	if activeRunCache != nil {
		return activeRunCache.types
	}
	if pendingTypes != nil {
		return pendingTypes.reuse
	}
	return nil
}

// pendingTypes is the types section of a run the run cache does not record, nil for every other run.
var pendingTypes *typesRecord

// typesRecord is the types section's own reading and writing, for a run with no recording to carry it: any
// flag, any named path, the editor's save of one file (#hfv0ae3). Those runs paid the full incremental
// session, about 180ms reading the build info and 150ms rewriting it, whether they checked 84 files or
// 3,889. The section is keyed per file on the shape fingerprint and on the compiler options, never on the
// arguments, so a scoped run may replay what a bare run recorded and the other way round.
type typesRecord struct {
	directory string
	anchor    program.PathAnchor
	reuse     *program.TypeDiagnosticsReuse

	// shapes is every project file's shape this run, kept for the next when computedShapes says any was
	// computed rather than carried from the table.
	shapes         map[string]program.SignatureEntry
	computedShapes int
	// shapesGraph is the graph shapes describe. See runCacheSession.shapesGraph.
	shapesGraph *program.Graph
}

// attachTypesCache gives a run the run cache does not record the types section, and the shapes it is keyed
// on. A recorded run already has both from attachFindingsCache.
func attachTypesCache(graph *program.Graph, location projectLocation) {
	if cacheOff || activeRunCache != nil || graph == nil {
		return
	}
	typesKey, err := typesCacheKey(graph, location)
	if err != nil {
		return
	}
	prepareCacheDirectory(location.Root)
	directory := cacheDirectory(location.Root)
	release, _ := holdTableReadLock(directory, tableWriterWait)
	anchor := program.NewPathAnchor(location.Root)
	table, _ := program.ReadCacheTable(directory, cacheTableIdentity(), program.CacheTableSections{Signatures: true, Types: true}, anchor)
	release()
	record := &typesRecord{directory: directory, anchor: anchor, reuse: program.NewTypeDiagnosticsReuse(table.Types, typesKey)}
	record.shapes, record.computedShapes = graph.Signatures(context.Background(), graph.SeedSignatures(table.Signatures))
	record.shapesGraph = graph
	graph.Shapes = record.shapes
	pendingTypes = record
}

// write puts the section, and the shapes when any was computed, into their files, and writes nothing when
// neither moved. The caller holds the table's lock.
func (record *typesRecord) write() {
	if record.reuse.Unchanged() && record.computedShapes == 0 {
		return
	}
	table := program.NewCacheTable()
	var sections program.CacheTableSections
	if recorded := record.reuse.Recorded(); recorded != nil {
		table.Types = recorded
		sections.Types = true
	}
	if record.computedShapes > 0 {
		table.Signatures = record.shapes
		sections.Signatures = true
	}
	if err := program.WriteCacheTable(record.directory, table, cacheTableIdentity(), sections, record.anchor); err != nil {
		cacheNote(record.directory, fmt.Sprintf("the cache table could not be written: %v", firstLine(err.Error())))
	}
}

// anyRuleKeyedOnShapes reports whether any rule declares rule.TypeReachShapes.
func anyRuleKeyedOnShapes(rules []rule.Rule) bool {
	for _, subject := range rules {
		if subject.TypeReach == rule.TypeReachShapes {
			return true
		}
	}
	return false
}

// findingsCacheKey covers everything a cacheable rule's answer depends on beyond a file's own bytes:
// the binary, the project root, the bytes of every file in the lint config's and the tsconfig's extends chains,
// and the rule set in order. Any change to one makes every entry a miss at once.
//
// The tsconfig chain is in it because it decides how a file parses, and a cacheable rule walks the
// parse. Rule options come from the lint config and the root, both already here.
func findingsCacheKey(graph *program.Graph, location projectLocation) (program.RunCacheKeyParts, error) {
	facts := []string{}
	// Every file in the lint config's extends chain, as the tsconfig's chain is below: a base edited
	// alone changes which rules run and with what options.
	// Under zero config they are the house sets and the project's own file when it has one.
	lintConfigFiles := houseSources(location)
	anchor := program.NewPathAnchor(location.Root)
	// An embedded set is hashed by its text, read the way the loader reads it, as a file is.
	for _, lintConfigFile := range lintConfigFiles {
		contents, err := configuration.SourceContents(lintConfigFile)
		if err != nil {
			return program.RunCacheKeyParts{}, err
		}
		facts = append(facts, fmt.Sprintf("lint-config %s=%x", anchor.Stable(lintConfigFile), sha256.Sum256(contents)))
	}
	// Under zero config which sets a file gets is decided by the whole program, so a file can change sets
	// without its own bytes changing, when another file starts importing next.
	if fingerprint := graph.LintConfig.DetectionFingerprint(); fingerprint != "" {
		facts = append(facts, "house-sets="+fingerprint)
	}
	compilerFacts, err := compilerOptionsFacts(graph, location)
	if err != nil {
		return program.RunCacheKeyParts{}, err
	}
	facts = append(facts, compilerFacts...)
	names := make([]string, 0, len(registry.All()))
	for _, registered := range registry.All() {
		names = append(names, registered.Name)
	}
	facts = append(facts, "rules="+strings.Join(names, "\x00"))
	return program.RunCacheKeyPartsOf(nil, anchor.Resolved(), facts...)
}

// typesCacheKey covers what a file's semantic diagnostics depend on that no shape fingerprint sees: the
// binary, the project root, and the compiler options, as the bytes of the tsconfig and every file it extends.
// The lint config and the rules are left out, since the compiler reads neither.
func typesCacheKey(graph *program.Graph, location projectLocation) ([sha256.Size]byte, error) {
	facts, err := compilerOptionsFacts(graph, location)
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	key, err := program.RunCacheKey(nil, program.NewPathAnchor(location.Root).Resolved(), facts...)
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	return sha256.Sum256([]byte(key)), nil
}

// compilerOptionsFacts is the tsconfig and every file in its extends chain, by path and bytes. It decides how
// each file parses, which a cacheable rule walks, and how each checks, which the types section replays.
func compilerOptionsFacts(graph *program.Graph, location projectLocation) ([]string, error) {
	facts := []string{}
	anchor := program.NewPathAnchor(location.Root)
	fileFact := func(label string, path string) error {
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		facts = append(facts, fmt.Sprintf("%s %s=%x", label, anchor.Stable(path), sha256.Sum256(contents)))
		return nil
	}
	if err := fileFact("tsconfig", location.ConfigFileName); err != nil {
		return nil, err
	}
	if graph.Config != nil {
		for _, extended := range graph.Config.ExtendedSourceFiles() {
			if err := fileFact("extends", extended.AsString()); err != nil {
				return nil, err
			}
		}
	}
	return facts, nil
}

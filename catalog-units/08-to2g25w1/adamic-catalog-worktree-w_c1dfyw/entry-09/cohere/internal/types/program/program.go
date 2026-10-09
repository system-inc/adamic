// Package program builds the type graph and hands the other phases a checker they call as a
// function.
//
// This is the claim the whole tool rests on. Standing in the checker's own language means
// `GetTypeAtLocation` is a function call rather than a subprocess: the tool this replaces spent
// 2.09s per run rebuilding a 9,530-file program in a separate process to answer four type
// questions, and 2.08s of a 2.25s run went to program construction rather than to rules.
//
// The program is built once, and every phase reads the same one.
package program

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/cohere/internal/lint/configuration"
)

// Graph is a built type graph: the program, and the checkers that answer questions about it.
//
// A Graph is never returned half-built. Build either hands back a Graph whose program resolved, or
// an error — there is no state in between, because the failure this package exists to make
// impossible is a program that silently contains nothing. An empty file list is indistinguishable
// from a clean tree, and that exact confusion let the gate this tool replaces print green over zero
// files for days.
type Graph struct {
	// Program is the compiler's own program handle. Exposed because a rule context carries it and
	// some type questions are asked of the program rather than of a checker.
	Program *compiler.Program

	// Config is the resolved tsconfig this graph was built from, kept so a later snapshot can hash
	// what it was built against rather than guess.
	Config *tsoptions.ParsedCommandLine

	// ConfigFileName is the absolute path of the tsconfig that produced this graph.
	ConfigFileName string

	// Anchor is the directory the graph was built in, which every fingerprint names paths relative to, so
	// that two spellings of one directory fingerprint every file alike (#547dhjz). See PathAnchor.
	Anchor PathAnchor

	// CompilerHost is the filesystem view the program was built through, kept because the
	// incremental machinery needs the same one.
	//
	// ReadBuildInfoProgram takes a host and a ParsedCommandLine, and it has to be the host the
	// program was built with rather than a fresh one: the build info's file paths are resolved
	// relative to it, and a second host with a different current directory or a cold cache
	// would resolve them somewhere else and report every file as changed. Silently, and in the
	// direction that looks like a cold cache rather than like a bug.
	CompilerHost compiler.CompilerHost

	// contentPack is the pack the program read its files through, and readIdentities the stats it read them
	// under when there was none. See ReadUnchanged.
	contentPack    *ContentPack
	readIdentities *readIdentities

	// LintConfig decides which rules apply to which files, and which files are not linted at all.
	//
	// Nil means every rule applies to every file, which is what the tests and any caller predating
	// this layer expect. The command always loads a real one and fails loudly if it cannot, so the
	// permissive default is reachable only from a caller that chose it deliberately.
	LintConfig *configuration.Config

	// RuleOptions says how to decode each rule's configuration, and which rules cannot run without
	// it. Empty means no rule takes options, which is what a test that did not set one expects.
	RuleOptions configuration.OptionsRegistry

	// RegisteredRuleNames is every rule the registry holds, whether or not this run selected it.
	//
	// The walk reads it to tell a suppression naming an unknown rule from one naming a real rule that
	// is merely not running here. Nil turns that check off, which is what a test that did not set it
	// expects: without a registry every name would read as unknown.
	RegisteredRuleNames []string

	// CollectTimings turns on per-rule cost measurement for the next Walk.
	//
	// Off by default and guarded at every timing site, so an ordinary run reads no clocks and
	// allocates no accumulators. The instrument should cost nothing when nobody asked for it.
	CollectTimings bool

	// FindingsReuse, when set, lets Walk serve unchanged files' cacheable findings from the last run
	// and records this run's for the next. Nil walks everything, exactly as before it existed.
	FindingsReuse *FindingsReuse

	// Readiness, when set, measures every walked file against cohere:adamic for the run's readiness, whatever
	// the project's chain enables. Nil measures nothing and the walk is exactly as before. See Readiness.
	Readiness *Readiness

	// Shapes is every project file's shape for this run (Graph.Signatures), when the caller computed them.
	// Walk keys shape-keyed rules on them; without them it keys those rules on the type fingerprint.
	Shapes map[string]SignatureEntry

	// projectFiles is the program's files that the tsconfig named, matched once by Build and verified
	// complete there. ProjectFiles returns it rather than recomputing, so the check that every named
	// file made it into the program happens where a failure can be returned instead of panicked.
	projectFiles []*ast.SourceFile

	// YieldedFiles is how many files the tsconfig named that a nearer one owns, left out of ProjectFiles.
	// See Options.Yielded.
	YieldedFiles int

	// typeGraphOnce and typeGraphParts hold typeGraph's answer, computed once: the type and the shape
	// fingerprints both read it, and the program it describes never changes after Build.
	typeGraphOnce  sync.Once
	typeGraphParts typeGraphParts

	// WalkOnForeignCheckers walks every file on a checker other than its own: a test instrument, for holding a
	// stolen file's findings and type diagnostics to its home checker's. See walkQueue.
	WalkOnForeignCheckers bool

	// ContentKeyedShapes is how many project files Signatures keyed on their content, with no recorded
	// declaration signature to compute theirs from. See Signatures.
	ContentKeyedShapes int

	// signatureEmits is how many files Signatures emitted declarations for, read by tests that prove when
	// it does and does not.
	signatureEmits int

	// FusedCheck, when set, has the next Walk type-check each file it visits, just before the file's rules run,
	// and the types phase read the result rather than check the program itself. Nil walks without checking.
	FusedCheck *FusedCheck

	// retired is set when the graph has been replaced, by Retire. See there.
	retired atomic.Bool
}

// Retire marks a graph replaced: the fix phase rewrote files and a new graph was built from the new bytes.
//
// A check still running on this graph is left to finish rather than cancelled, and what it finds is never
// recorded into the run cache's types section, since it describes bytes that no longer exist. It is not
// cancelled because the compiler's whole-program check cannot be: a checker that sees its context end
// marks itself cancelled, the compiler then hands it the next file in its group, and GetGlobalDiagnostics
// panics on a compiler goroutine nothing here can recover. The gate itself died that way, a bare `cohere`
// in 5 of 6 runs once its fix phase rewrote a file before the early check finished (#sm79kfv).
func (g *Graph) Retire() {
	g.retired.Store(true)
}

// Retired reports whether Retire was called.
func (g *Graph) Retired() bool {
	return g.retired.Load()
}

// Options controls how a graph is built.
type Options struct {
	// ConfigFileName is the tsconfig to build from. Relative paths resolve against CurrentDirectory.
	ConfigFileName string

	// CurrentDirectory is the working directory paths resolve against. Empty means the process's own.
	CurrentDirectory string

	// LibraryPath overrides where the lib.*.d.ts files come from.
	//
	// Empty means the ones embedded in our pinned typescript-go, which is the correct default: the
	// checker and its libs are one version, and mixing a newer lib into an older checker produces
	// hundreds of spurious TS2550/TS2554 (measured: 118 of them on the ahra tree). Point this at a
	// specific lib directory only when deliberately reproducing another toolchain's result.
	LibraryPath string

	// Checkers is how many checkers to create. Zero means the compiler's own default.
	//
	// It is the count that parallelizes the checking phase, not the graph build: construction is
	// parse and resolve, which this does not touch. Raising it trades memory for wall clock, since
	// each checker holds its own type tables.
	Checkers int

	// SingleThreaded forces one checker instead of one per core. Off by default; useful when a
	// measurement needs to be reproducible, or when a caller wants types from different files to be
	// comparable (types from two checkers cannot be mixed).
	SingleThreaded bool

	// Overlay is file contents that stand in for the disk's, by absolute path: an editor's unsaved
	// buffer, so the graph describes the text about to be saved rather than the text being replaced.
	// Nil reads everything from disk.
	Overlay map[string]string

	// Inputs, when set, is told every path the build reads or looks for on disk, so the run cache can
	// record what this build depended on by observation rather than by a list. Nil records nothing and
	// costs nothing.
	Inputs *InputRecorder

	// ContentPack, when set, serves files whose stat is unchanged from the bytes recorded by earlier runs,
	// so they are not opened, and collects what this build read from disk. Nil reads everything from disk.
	ContentPack *ContentPack

	// LibraryParses, when set, serves the bundled lib files from one parse shared by every program built
	// with it, rather than parsing them for each program. Nil parses them per program, as before. See
	// LibraryParses.
	LibraryParses *LibraryParses

	// Timing, when set, is filled with what the build cost by part. Nil reads no clocks and wraps
	// nothing, so a build nobody is timing costs what it did before this existed. See GraphTiming.
	Timing *GraphTiming

	// FileSystem, when set, is the disk the build reads instead of the real one: a MemoryFS, for a test
	// that builds its fixture from a map. The bundled lib files are laid over it as over the real disk.
	// Nil reads the real disk, as before.
	FileSystem vfs.FS

	// Yielded is files this tsconfig includes that a nearer one owns, by absolute path: they stay in the
	// program, so every type question about them is answered, and leave ProjectFiles, so no phase lints
	// them or reports their diagnostics here. Nil yields nothing. See ReadProjectConfig, and discovery's
	// ownership in the command.
	Yielded map[string]struct{}

	// CheckedStats, when set, answers the build's existence checks and stats from what the run cache's check found
	// at each path it statted this run, rather than asking the disk again. Beneath the input recorder, so a run
	// cache still records every path the build asked about. Nil asks the disk. See StatSnapshot.
	CheckedStats *StatSnapshot

	// Listings, when set, serve the build's directory listings from ones already read, discovery's, rather
	// than reading those directories again. Beneath the input recorder, so a run cache still records every
	// listing the build asked for. Nil reads every listing from disk. See DirectoryListings.
	Listings *DirectoryListings
}

// Build resolves a tsconfig and constructs the program and its checkers.
//
// Every failure here is loud. A tsconfig that does not exist, one that does not parse, one whose
// options are contradictory, and one that matches no files are four different errors rather than
// four flavors of an empty result, because the whole point of this package is that a run which
// checked nothing must not be able to look like a run that found nothing.
func Build(options Options) (*Graph, error) {
	graph, err := buildOnce(options)
	if err != nil {
		return nil, err
	}

	// The config's file list and the program's loaded files are two reads of the tree, a moment
	// apart. A file deleted or replaced between them is listed but cannot be loaded, which is routine
	// when many writers share the tree: an atomic save, a throwaway probe. That is a moved tree, not a
	// broken config, so it gets one fresh build, from a fresh filesystem cache, before anything is
	// reported. Measured on 2026-10-03: with a root listed and unloadable, the guard this replaced
	// panicked and blamed path casing.
	verdict, missing := graph.timedVerify(options)
	if verdict == rootsMoved {
		graph, err = buildOnce(options)
		if err != nil {
			return nil, err
		}
		verdict, missing = graph.timedVerify(options)
		if verdict == rootsMoved {
			return nil, fmt.Errorf("program: files %s named disappeared before they could be read, in two "+
				"builds in a row: %s. Something is rewriting the tree faster than a build; run again once "+
				"it settles", graph.ConfigFileName, describeRoots(missing))
		}
	}

	switch verdict {
	case rootsUnreadable:
		return nil, fmt.Errorf("program: %d of the files %s named exist but could not be read, so the "+
			"program left them out: %s. Linting the rest would silently skip them",
			len(missing), graph.ConfigFileName, describeRoots(missing))
	case rootsUnmatched:
		return nil, errors.New(projectFilesMismatchMessage(len(graph.Config.FileNames()), len(graph.projectFiles)))
	}
	// After verifying, which holds the program to every file the config named, yielded ones included.
	graph.yield(options.Yielded)
	return graph, nil
}

// yield takes the files a nearer tsconfig owns out of ProjectFiles, and counts them.
func (g *Graph) yield(yielded map[string]struct{}) {
	if len(yielded) == 0 {
		return
	}
	kept := g.projectFiles[:0:0]
	for _, sourceFile := range g.projectFiles {
		if _, isYielded := yielded[filepath.Clean(filepath.FromSlash(sourceFile.FileName().AsString()))]; isYielded {
			g.YieldedFiles++
			continue
		}
		kept = append(kept, sourceFile)
	}
	g.projectFiles = kept
}

// timedVerify is verifyProjectFiles against the disk, timed when the build is. A build over another
// filesystem asks that one, since the real disk knows nothing of its files.
func (g *Graph) timedVerify(options Options) (rootsVerdict, []string) {
	inspect := inspectRootOnDisk
	if options.FileSystem != nil {
		inspect = func(fileName string) rootState {
			if options.FileSystem.FileExists(tspath.RootedFilePath(fileName)) {
				return rootReadable
			}
			return rootAbsent
		}
	}
	timing := options.Timing
	if timing == nil {
		return g.verifyProjectFiles(inspect)
	}
	started := time.Now()
	verdict, missing := g.verifyProjectFiles(inspect)
	timing.Verify += time.Since(started)
	return verdict, missing
}

// rootState is what the disk says about a file the config named, read now rather than from the build's
// cache, which still remembers the tree as it was when the config was globbed.
type rootState int

const (
	rootAbsent rootState = iota
	rootUnreadable
	rootReadable
)

// rootsVerdict is what a build's unloaded roots mean.
type rootsVerdict int

const (
	// rootsComplete: every file the config named is in the program.
	rootsComplete rootsVerdict = iota

	// rootsMoved: every missing root is gone from disk, so the tree changed between the two reads.
	rootsMoved

	// rootsUnreadable: a missing root exists but cannot be read.
	rootsUnreadable

	// rootsUnmatched: a missing root exists and is readable, so the program loaded it under a path the
	// config's name does not canonicalize to. This is the casing defect the original guard was for.
	rootsUnmatched
)

// classifyMissingRoots decides what a set of named-but-unloaded roots means.
//
// A pure function of what the disk says, so every verdict is tested directly. The order is the order
// of seriousness: one readable root that did not match is a real defect however many others vanished,
// so it wins over a moved tree, and an unreadable root wins over one that vanished.
func classifyMissingRoots(missing []string, inspect func(string) rootState) rootsVerdict {
	if len(missing) == 0 {
		return rootsComplete
	}
	verdict := rootsMoved
	for _, fileName := range missing {
		switch inspect(fileName) {
		case rootReadable:
			return rootsUnmatched
		case rootUnreadable:
			verdict = rootsUnreadable
		}
	}
	return verdict
}

// verifyProjectFiles matches the config's files to the program's, keeps the matches for ProjectFiles,
// and says what any missing ones mean.
func (g *Graph) verifyProjectFiles(inspect func(string) rootState) (rootsVerdict, []string) {
	rootPaths := make(map[tspath.PathKey]string, len(g.Config.FileNames()))
	for _, fileName := range g.Config.FileNames() {
		rootPaths[toPath(fileName.AsString(), g.Anchor.Root(), g.Config.CaseSensitivity())] = fileName.AsString()
	}

	files := make([]*ast.SourceFile, 0, len(rootPaths))
	loaded := make(map[tspath.PathKey]struct{}, len(rootPaths))
	for _, sourceFile := range g.Program.GetSourceFiles() {
		if _, isRoot := rootPaths[sourceFile.PathKey()]; isRoot {
			files = append(files, sourceFile)
			loaded[sourceFile.PathKey()] = struct{}{}
		}
	}
	g.projectFiles = files

	missing := []string{}
	for path, fileName := range rootPaths {
		if _, isLoaded := loaded[path]; !isLoaded {
			missing = append(missing, fileName)
		}
	}
	sort.Strings(missing)
	return classifyMissingRoots(missing, inspect), missing
}

// inspectRootOnDisk asks the real disk, past every cache, what is at a path now.
func inspectRootOnDisk(fileName string) rootState {
	if _, err := os.Stat(fileName); err != nil {
		return rootAbsent
	}
	file, err := os.Open(fileName)
	if err != nil {
		return rootUnreadable
	}
	file.Close()
	return rootReadable
}

// describeRoots names the first few missing roots, so an error points at files rather than a count.
func describeRoots(fileNames []string) string {
	const shown = 3
	if len(fileNames) <= shown {
		return strings.Join(fileNames, ", ")
	}
	return fmt.Sprintf("%s and %d more", strings.Join(fileNames[:shown], ", "), len(fileNames)-shown)
}

// buildOnce resolves the config and builds the program from it, once, without checking that every
// named file was loaded. Build does that, because only it can decide to build again.
func buildOnce(options Options) (*Graph, error) {
	currentDirectory := options.CurrentDirectory
	if currentDirectory == "" {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("resolving the current directory: %w", err)
		}
		currentDirectory = workingDirectory
	}
	currentDirectory = tspath.NormalizePath(currentDirectory)

	if options.ConfigFileName == "" {
		return nil, errors.New("no tsconfig given: a type graph needs a config to know what belongs in it")
	}
	configFileName := options.ConfigFileName
	if !tspath.IsRootedDiskPath(configFileName) {
		configFileName = tspath.CombinePaths(currentDirectory, configFileName)
	}
	configFileName = tspath.NormalizePath(configFileName)

	// bundled.WrapFS overlays the embedded lib.*.d.ts files onto the real disk, so `lib.dom.d.ts`
	// resolves without anything being installed. cachedvfs memoizes stat and readdir: config
	// resolution and module resolution ask the same directories about the same files repeatedly, and
	// on a 9,530-file program that repetition is most of the syscall traffic.
	//
	// The input recorder wraps the real disk innermost: beneath the memoizing layer, so it sees each path
	// about once, and beneath the lib overlay, so the embedded libs never reach it.
	//
	// The content pack is beneath even that, so a file it serves is still a file the recorder saw read, and so
	// are the run cache's check's answers, so a path they answer is still a path the recorder saw asked about.
	var disk vfs.FS = osvfs.FS()
	if options.FileSystem != nil {
		disk = options.FileSystem
	}
	if options.Listings != nil {
		disk = &listingFS{FS: disk, listings: options.Listings}
	}
	// Without a pack, the program still records the stat each file was read under, so ReadUnchanged can answer.
	var identities *readIdentities
	if options.ContentPack != nil {
		disk = options.ContentPack.wrap(disk)
	} else {
		identities = &readIdentities{byPath: map[string]fileIdentity{}}
		disk = &identityFS{FS: disk, identities: identities}
	}
	var checkedAnswers atomic.Int64
	if options.CheckedStats != nil {
		disk = &checkedStatsFS{FS: disk, snapshot: options.CheckedStats, answered: &checkedAnswers}
	}
	if options.Inputs != nil {
		disk = &recordingFS{FS: disk, recorder: options.Inputs}
	}
	// Timed just beneath the cache, so it counts the calls that reach the disk (or the pack), and what the
	// recorder costs them, but not the ones the cache answered.
	var timedDisk *timingFS
	var configDisk, programDisk diskCounters
	if options.Timing != nil {
		options.Timing.Builds++
		timedDisk = &timingFS{FS: disk}
		timedDisk.current.Store(&configDisk)
		disk = timedDisk
	}
	var fileSystem vfs.FS = cachedvfs.From(bundled.WrapFS(disk))
	if len(options.Overlay) > 0 {
		fileSystem = newOverlayFS(fileSystem, options.Overlay)
	}

	configStarted := time.Now()
	if !fileSystem.FileExists(tspath.RootedFilePath(configFileName)) {
		return nil, fmt.Errorf("no tsconfig at %s", configFileName)
	}

	// A nil extended-config cache is supported upstream and correct for a one-shot build: the cache
	// only pays off across repeated parses of the same `extends` chain within one process.
	config, configErrors := tsoptions.GetParsedCommandLineOfConfigFile(tspath.RootedFilePath(configFileName), &core.CompilerOptions{}, nil, fileSystem, nil)
	if len(configErrors) > 0 {
		return nil, fmt.Errorf("reading %s: %w", configFileName, joinDiagnostics(configErrors))
	}
	if config == nil {
		return nil, fmt.Errorf("reading %s: the config parsed to nothing", configFileName)
	}

	// A malformed option arrives on the config rather than in the returned error slice, and the two
	// channels do not overlap: an invalid `target` produces zero returned errors and one
	// GetConfigFileParsingDiagnostics entry. Checking only the slice — which is what reading the
	// signature suggests — means a config with a typo builds a program against silently wrong options
	// and reports whatever that produces as truth.
	//
	// Except a diagnostic about one option's name or value, an option TypeScript 7 no longer knows above
	// all: the graph builds over it, and it is reported with the program's own diagnostics (they include
	// the config's), so the options are never silently wrong. See configParsingOptionCodes.
	if configFileDiagnostics := config.GetConfigFileParsingDiagnostics(); len(configFileDiagnostics) > 0 && !onlyOptionValueDiagnostics(configFileDiagnostics) {
		return nil, fmt.Errorf("reading %s: %w", configFileName, joinDiagnostics(configFileDiagnostics))
	}

	// The config resolving to zero files is a configuration error, not a clean tree. Catching it here
	// means the caller never has to distinguish "nothing to check" from "nothing was checked", which
	// is the distinction every silent-green failure in this project's history collapsed.
	if len(config.FileNames()) == 0 {
		return nil, fmt.Errorf("%s matched no files: check its include, files, and exclude", configFileName)
	}
	if options.Timing != nil {
		options.Timing.Config += time.Since(configStarted)
		options.Timing.ConfigDisk = options.Timing.ConfigDisk.plus(configDisk.snapshot())
		timedDisk.current.Store(&programDisk)
	}

	libraryPath := options.LibraryPath
	if libraryPath == "" {
		libraryPath = bundled.LibPath().AsString()
	}
	// The last three arguments arrived with the move to `microsoft/TypeScript`. Nil is correct for all
	// three here: the extended-config cache only pays off across repeated parses of the same `extends`
	// chain within one process, `trace` is nil-guarded upstream and we have nowhere to route compiler
	// tracing, and the content mapper serves the language server rather than a one-shot check. The host
	// takes no current directory: since typed file paths (#64159) every name it is given is rooted, and
	// the program answers GetCurrentDirectory with its config's directory.
	compilerHost := compiler.NewCachedFSCompilerHost(fileSystem, tspath.RootedDirectoryPath(libraryPath), nil, nil, nil)

	singleThreaded := core.TSUnknown
	if options.SingleThreaded {
		singleThreaded = core.TSTrue
	}

	// The count goes onto the compiler options rather than being tracked beside them, so the compiler
	// and Workers() read the same number from the same place. A count kept alongside is a count that
	// can drift from the one the checkers were actually built with.
	checkerCount := options.Checkers
	if checkerCount <= 0 {
		checkerCount = defaultCheckerCount()
	}
	config.CompilerOptions().Checkers = &checkerCount

	// The lib files TypeScript ships are never type-checked here, whatever the tsconfig says. A diagnostic
	// in one is never reported, since a run reports the files the tsconfig names and those are never
	// among them, so checking them was work whose output was thrown away: on a one-file program, 84ms of
	// a 88ms types phase and 0.10s of 0.12s user CPU, every cold run of every small project and about 260
	// launches of the test suite (#fay5rd1). A diagnostic about a file the tsconfig names comes from
	// checking that file, and the globals every checker merges at startup are merged either way.
	//
	// The incremental build info compares the option only to decide whether the lib files' diagnostics
	// can be copied, so a project's own tsc or tsgo sharing that build info rechecks the lib files after a
	// cohere run and nothing else.
	config.CompilerOptions().SkipDefaultLibCheck = core.TSTrue

	// JSDocParsingMode used to be set here to parse JSDoc only where it can carry types. The option
	// was removed from ProgramOptions in the move to `microsoft/TypeScript` — the compiler now decides
	// per file rather than taking a program-wide mode — so there is nothing to pass and nothing to
	// preserve.
	// With shared lib parses, every load goes through the sharing host, the graph's included, so anything
	// that loads a file later shares as the build did.
	if options.LibraryParses != nil {
		compilerHost = &libraryParsesHost{CompilerHost: compilerHost, parses: options.LibraryParses}
	}

	// The program is built through a timing host when the build is timed, and the graph keeps the plain
	// one either way, so nothing after the build reads through the instrument.
	var programHost compiler.CompilerHost = compilerHost
	var timedHost *timingHost
	if options.Timing != nil {
		timedHost = &timingHost{CompilerHost: compilerHost}
		programHost = timedHost
	}
	programStarted := time.Now()
	builtProgram := compiler.NewProgram(compiler.ProgramOptions{
		Config:         config,
		Host:           programHost,
		SingleThreaded: singleThreaded,
	})
	if options.Timing != nil {
		options.Timing.Program += time.Since(programStarted)
		options.Timing.SourceFileLoads += timedHost.loads.Load()
		options.Timing.SourceFileSummed += time.Duration(timedHost.summed.Load())
		options.Timing.ProgramDisk = options.Timing.ProgramDisk.plus(programDisk.snapshot())
		if options.ContentPack != nil {
			// The pack's counts are cumulative over its life, which is this build and any rebuild before it,
			// so they are read rather than added.
			options.Timing.PackServed, options.Timing.PackRead, _ = options.ContentPack.Counts()
		}
		options.Timing.CheckedAnswers += checkedAnswers.Load()
	}
	if builtProgram == nil {
		return nil, fmt.Errorf("building a program from %s produced nothing", configFileName)
	}

	// The same guard one layer down, and it is not redundant with the one above: the config can match
	// files that the program then fails to load. Both are loud, and neither is an empty success.
	if len(builtProgram.GetSourceFiles()) == 0 {
		return nil, fmt.Errorf("the program built from %s contains no source files", configFileName)
	}

	return &Graph{
		Program:        builtProgram,
		Config:         config,
		ConfigFileName: configFileName,
		CompilerHost:   compilerHost,
		Anchor:         NewPathAnchor(currentDirectory),
		contentPack:    options.ContentPack,
		readIdentities: identities,
	}, nil
}

/*
 * ReadUnchanged reports whether text, the program's copy of a file, is the file's bytes now, so a reader can
 * use it instead of reading the file again (#q6dey77). Exact, by three checks:
 *   - the program read the file under a stat, its content pack's or its own, and a stat taken now matches it,
 *     so the file holds the bytes the program read;
 *   - those bytes were the program's copy exactly. The program reads through typescript-go's decodeBytes,
 *     which drops a UTF-8 byte order mark and decodes UTF-16 behind its mark, and returns every other file
 *     byte for byte. So the file must not start with a mark (#hkv1hgp). The recorded size must also be the
 *     copy's length, which no dropped mark leaves true and which costs no read, so a marked file is turned
 *     away before it is opened. The length alone isn't enough: UTF-16 heavy in characters that take three
 *     bytes in UTF-8 can decode to exactly its own size.
 * False on any doubt, so the reader reads the disk, as it would have.
 */
func (g *Graph) ReadUnchanged(path string, text string) bool {
	if g == nil {
		return false
	}
	var identity fileIdentity
	var known bool
	if g.contentPack != nil {
		identity, known = g.contentPack.readIdentity(path)
	} else {
		identity, known = g.readIdentities.identity(path)
	}
	if !known || identity.size != int64(len(text)) {
		return false
	}
	if current, statted := statIdentity(path); !statted || current != identity {
		return false
	}
	return !startsWithByteOrderMark(path)
}

// SourceFiles is every file in the program, third-party declarations included.
//
// This is the program's population, not the set a phase should lint. Use ProjectFiles for that.
func (g *Graph) SourceFiles() []*ast.SourceFile {
	return g.Program.GetSourceFiles()
}

// ProjectFiles is the subset of the program that the tsconfig actually named — our code, without
// the declarations it pulled in.
//
// On the ahra tree that is roughly 3,400 files out of 9,530. The other 6,100 are `node_modules`
// `.d.ts` that the checker needs in order to answer questions and that no rule should visit: a rule
// walking them would cost most of the run to report findings against code nobody here can edit.
func (g *Graph) ProjectFiles() []*ast.SourceFile {
	// Matched and verified once, by Build. A lookup that misses reports the same empty result as a
	// config that named nothing, and this one has missed twice: the compiler lowercases Path() on a
	// case-insensitive filesystem while the config keeps its casing (matched zero of 3,407), and a file
	// can be deleted between the config's glob and the program's read (matched 3,789 of 3,790, with many
	// writers sharing the tree). Build tells the two apart and fails on the first, so nothing here can
	// return a silently partial list.
	return g.projectFiles
}

// projectFilesMismatchMessage says what went wrong and where to look.
//
// Split out so a test can assert the wording. A panic reading only "mismatch" sends the next reader
// to the tsconfig, which is the one place the fault is not.
func projectFilesMismatchMessage(named int, matched int) string {
	return fmt.Sprintf(
		"program: the tsconfig named %d files but only %d matched the program's own paths. "+
			"This is a path-canonicalization mismatch, not an empty project: the compiler lowercases "+
			"Path() on a case-insensitive filesystem while the config keeps its original casing. "+
			"Linting the %d that matched would silently skip the other %d.",
		named, matched, matched, named-matched,
	)
}

// CheckerForFile returns the checker that owns a file, and a function to release it.
//
// There are several checkers, one per worker, and a file belongs to exactly one of them. That
// ownership is not an optimization detail a caller may ignore: types obtained from two different
// checkers cannot be compared or mixed, so a rule must ask the checker that owns the file it is
// looking at. Asking for the wrong one produces answers that are individually plausible and jointly
// meaningless, which is the worst failure shape available here.
//
// The release function must be called, conventionally by defer.
func (g *Graph) CheckerForFile(ctx context.Context, sourceFile *ast.SourceFile) (*checker.Checker, func()) {
	return g.Program.GetTypeCheckerForFileExclusive(ctx, sourceFile)
}

// Diagnostics returns TypeScript's own findings for one file: syntax, then binding, then semantics.
//
// The order is the one `tsc` reports in and it matters. A file that does not parse produces
// cascading nonsense from the later phases, so when syntax fails we stop rather than bury the real
// error under fifty invented ones.
//
// Prefer AllDiagnostics for a whole-program check. Calling this in a loop over every file is more
// than three times slower for a reason that is invisible at the call site — see the note there.
func (g *Graph) Diagnostics(ctx context.Context, sourceFile *ast.SourceFile) []*ast.Diagnostic {
	syntactic := g.Program.GetSyntacticDiagnostics(ctx, sourceFile)
	if len(syntactic) > 0 {
		return syntactic
	}

	diagnostics := g.Program.GetBindDiagnostics(ctx, sourceFile)
	return append(diagnostics, g.Program.GetSemanticDiagnostics(ctx, sourceFile)...)
}

// AllDiagnostics returns TypeScript's findings for the whole program.
//
// This is not a convenience wrapper around Diagnostics in a loop, and the difference is a measured
// 3.7s versus 1.1s on a 3,408-file tree. The compiler's diagnostic entry points branch on whether a
// file was named: pass one and it checks that file lazily, on whichever single checker owns it; pass
// nil and it first runs CheckSourceFiles, which fans every checker out across all four workers at
// once. So a loop that asks file by file quietly serializes the most expensive phase in the tool
// onto one core, and it does it without any call looking wrong.
//
// The findings are identical either way. Only the wall clock differs, which is exactly what makes
// this the kind of mistake that survives review.
func (g *Graph) AllDiagnostics(ctx context.Context) []*ast.Diagnostic {
	return g.AllDiagnosticParts(ctx).All()
}

// AllDiagnosticParts is AllDiagnostics with its parts kept apart. See TypeDiagnosticParts.
func (g *Graph) AllDiagnosticParts(ctx context.Context) TypeDiagnosticParts {
	syntactic := g.Program.GetSyntacticDiagnostics(ctx, nil)
	if len(syntactic) > 0 {
		return TypeDiagnosticParts{Syntactic: syntactic}
	}
	return TypeDiagnosticParts{Bind: g.Program.GetBindDiagnostics(ctx, nil), Semantic: g.Program.GetSemanticDiagnostics(ctx, nil)}
}

// ConfigDiagnostics returns findings about the configuration itself rather than about any file —
// contradictory options, an unreachable lib, a target that does not exist.
func (g *Graph) ConfigDiagnostics(ctx context.Context) []*ast.Diagnostic {
	// GetOptionsDiagnostics was folded into GetProgramDiagnostics upstream: an option that contradicts
	// another is now reported with the rest of the program's own findings rather than through its own
	// accessor. The set is the same; only the door changed.
	diagnostics := g.Program.GetConfigFileParsingDiagnostics()
	diagnostics = append(diagnostics, g.Program.GetProgramDiagnostics()...)
	diagnostics = append(diagnostics, g.Program.GetGlobalDiagnostics(ctx)...)
	return diagnostics
}

// Workers is how many files can be checked at once.
//
// It matches the checker count the program built, because a file must be visited by the checker that
// owns it and there is no benefit to more goroutines than there are checkers to serve them.
//
// This asks the program rather than remembering what we requested. Until the move to
// `microsoft/TypeScript`, `Program.SingleThreaded()` was unexported and the checker count was a
// private constant, so this had to mirror the compiler's own arithmetic and hope the two stayed in
// step. Both are reachable now, so the mirror is gone: a count that disagreed with the compiler's
// would hand files to a worker whose checker does not own them, and types from two checkers cannot
// be mixed.
func (g *Graph) Workers() int {
	if g.Program.SingleThreaded() {
		return 1
	}

	checkerCount := defaultCheckerCount()
	if requested := g.Program.Options().Checkers; requested != nil {
		checkerCount = *requested
	}

	// The clamp is upstream's, reproduced rather than approximated, because agreeing with the
	// compiler is the entire point of this function. It caps at the file count for the obvious reason
	// that a fifth checker on a four-file program has nothing to check, and at 256 as a ceiling.
	//
	// Without it a config asking for 100 checkers on a 64-file program gets 64 checkers upstream and
	// 100 workers here, and the 36 extra stride over files whose owning checker is a different
	// instance than the one they would be handed. Types from two checkers cannot be mixed, so that is
	// wrong answers rather than wasted goroutines. Measured: Workers() returned 100 against
	// upstream's 64 before this.
	return max(min(checkerCount, len(g.Program.GetSourceFiles()), 256), 1)
}

// defaultCheckerCount is how many checkers we ask for when nobody says otherwise.
//
// The compiler's own default is a fixed 4, which was chosen for a machine we are not on: on a
// 16-core box it leaves the checking phase at a 2.6x speedup where the hardware allows more, so the
// ceiling was the setting rather than the workload. Measured on the ahra tree, 9,982 files:
//
//	checkers   types    allocated
//	       4   942ms      2,150MB
//	       8   758ms      2,483MB
//	      16   686ms      2,944MB
//	      32   762ms      3,467MB
//
// The curve turns over after core count, so this tracks GOMAXPROCS rather than hardcoding 16: a
// fixed 16 would oversubscribe a 4-core machine for the same reason a fixed 4 undersubscribes this
// one. The diagnostic count was 2 at every setting, which is the part that makes the speed
// trustworthy — a checking phase that got faster by checking less reports fewer findings, and this
// one does not.
//
// Memory is the trade and it is real: 37% more allocated between 4 and 16. That is affordable on a
// developer machine and worth watching if this ever runs somewhere small.
//
// And it stops at checkerCeiling, where the whole run's wall stops improving. Measured 2026-10-05 on this
// 16-core machine (#xwv641q): a cold ahra run, pinned at 802d78bb, on cache's interleaved pairs, medians
// of 4 rounds at load about 4 to 8:
//
//	checkers   wall    user     sys     allocated   objects
//	       8   2.23s   16.58s   3.77s   10.15 GB    112.8M
//	      10   2.02s   18.03s   4.12s   10.29 GB    113.3M
//	      12   1.93s   18.75s   4.22s   10.37 GB    113.6M
//	      14   1.94s   18.94s   4.24s   10.49 GB    113.9M
//	      16   1.92s   19.28s   4.24s   10.64 GB    114.4M
//
// Past 12 a checker buys no wall and costs CPU and memory, and a profile pair put that cost in the runtime
// acquiring memory (madvise, page zeroing), not in checkers repeating each other's types. A machine with
// 12 cores or fewer, most laptops, runs one per core as before; whether their knee sits lower is
// unmeasured.
func defaultCheckerCount() int {
	return checkerCountFor(runtime.GOMAXPROCS(0))
}

// checkerCeiling is the most checkers a build asks for unasked. See defaultCheckerCount.
const checkerCeiling = 12

// checkerCountFor is the default checker count for a machine with this many cores: one per core, from one
// to checkerCeiling.
func checkerCountFor(cores int) int {
	return min(max(cores, 1), checkerCeiling)
}

// toPath normalizes a file name the way the compiler keys its file table, so a lookup by path finds
// the file the compiler stored rather than a near-miss that differs only in case or separators.
func toPath(fileName string, currentDirectory string, caseSensitivity tspath.CaseSensitivity) tspath.PathKey {
	return caseSensitivity.PathKey(tspath.RootedPath(tspath.GetNormalizedAbsolutePath(fileName, tspath.RootedDirectoryPath(currentDirectory))))
}

// joinDiagnostics turns compiler diagnostics into one error, capped so a config with a hundred
// problems reports the first few rather than a wall.
func joinDiagnostics(diagnostics []*ast.Diagnostic) error {
	const shown = 5

	messages := make([]error, 0, min(len(diagnostics), shown)+1)
	for index, diagnostic := range diagnostics {
		if index == shown {
			messages = append(messages, fmt.Errorf("and %d more", len(diagnostics)-shown))
			break
		}
		messages = append(messages, fmt.Errorf("TS%d: %s", diagnostic.Code(), diagnostic.MessageText()))
	}
	return errors.Join(messages...)
}

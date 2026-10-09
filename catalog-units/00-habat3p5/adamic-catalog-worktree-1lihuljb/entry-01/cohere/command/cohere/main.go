// Command cohere type-checks, lints, fixes, and formats a TypeScript codebase in one process,
// over one AST, against one type graph.
//
// See the domain body at `ahra tasks show system_cohere` for why this exists and what it replaces.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/locale"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/gitignore"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/release/packaging"
	"github.com/system-inc/cohere/internal/types/program"
	"github.com/system-inc/cohere/internal/unused_exports"
)

// cacheOff is `--no-cache`: no cache is read or written for the project. Package-level because the
// run cache, the format record and the types phase each consult it, and it has one writer, the flag.
var cacheOff bool

// processStart is stamped before anything else runs, so the phase line can say how much of the run
// its own numbers explain. A package-level variable rather than a parameter because `run` already
// takes none, and the value has exactly one writer, at initialization.
var processStart = time.Now()

func main() {
	if err := run(); err != nil {
		// A failure of the tool is never recorded by the run cache, and the recording stops before the
		// error prints so the message reaches the terminal rather than a pipe nobody is reading.
		abandonRunCache()
		fmt.Fprintf(os.Stderr, "cohere: %v\n", err)
		exitProcess(1)
	}
	finishRunCache(0)
}

func run() error {
	// A subcommand is matched before flag.Parse, because the flag package stops at the first
	// argument it does not recognize and would hand `rename` back as a positional while
	// swallowing the flags meant for it. The dispatcher forwards unknown arguments untouched, so the
	// verb arrives here intact.
	//
	// This is a verb rather than a flag on the pipeline for the reason set out on runRenameVerb: a
	// rename is imperative where every other write here is reactive, and it must not inherit the
	// default-on mutation that `--fix` correctly has.
	if isRenameVerb(os.Args[1:]) {
		return runRenameVerb(os.Args[2:])
	}

	configFileName := flag.String("tsconfig", projectMarker,
		"the tsconfig that defines the program, relative to --directory or else to where you typed it; unnamed, the file of this name in --directory, or else the nearest one at or above the working directory")
	directory := flag.String("directory", "",
		"the project root, which every relative path resolves against (default: the directory of the nearest tsconfig.json or Package.swift at or above the working directory)")
	typesOnly := flag.Bool("types", false, "report TypeScript's own diagnostics only, running no rules and writing no fixes")
	lintOnly := flag.Bool("lint", false, "run the lint rules only, reporting their findings without fixing them and without TypeScript's diagnostics")
	lintConfigFileName := flag.String("lint-config", "CohereSettings.json",
		"the settings file that says which rules apply to which files, relative to --directory or else to where you typed it; unnamed, the file of this name at the project root")
	singleThreaded := flag.Bool("single-threaded", false, "use one checker instead of several")
	// The whole mutating phase, fixes and format (Kirk's ruling, 2026-10-04, reversing b617c98b's fix-only
	// reading), narrowed from the rest of the run; --no-format leaves formatting out.
	fixOnly := flag.Bool("fix", false, "apply fixes and format, running no other phase (--no-format to leave formatting out)")
	// The promise is about the project's source, and it is stated with its boundary because two writes sit
	// outside it on purpose. cohere keeps its cache for the project in `<root>/.cache/cohere/`, which is its
	// own and which `--no-cache` turns off. And the launcher rebuilds cohere itself when its rules
	// changed, into the gitignored `.cache/cohere/` of the cohere checkout; withholding that would run a
	// binary that does not match the rules on disk. See recordDevelopmentHash in internal/release/dispatch.
	noFix := flag.Bool("no-fix", false,
		"mutate no source in the checked project: report what would change, fixes and formatting both, without "+
			"writing a byte of it, and exit nonzero if anything would (cohere still keeps its own cache in the "+
			"project's .cache/cohere, unless --no-cache)")
	// The opt-out from formatting by default, for a run that wants fixes, types and lint alone.
	noFormat := flag.Bool("no-format", false,
		"leave formatting out of a bare run, --fix or --no-fix: fix, type-check and lint only")
	formatAll := flag.Bool("format-all", false, "format every file, not only the ones not on record as formatted (implies --format)")
	// The fix phase narrowed to formatting, the way --types and --lint narrow the run to theirs, for a gate
	// whose question is formatting alone: a lint finding has no say in its exit.
	formatOnly := flag.Bool("format-only", false,
		"format only, proposing no fixes and running no other phase (implies --format); with --no-fix, the commit gate's "+
			"format check: every file formatting would change here or in a nested repository is a finding, and so is "+
			"one the formatter could not read, and the exit is 0 only when there are none")
	// A cold run on purpose: for measuring one, and for anyone who suspects a cache. Every cache cohere
	// keeps for a project is named here, so a cold number cannot be read as a warm one: the cache table
	// (the run replay, the per-file findings, the signatures, the format record) and the tsconfig's
	// incremental build info.
	noCache := flag.Bool("no-cache", false,
		"read nothing from and write nothing to this project's caches (the cache table and the tsconfig's incremental build info), so every phase computes from source")
	// A binary that implements no rule and a binary whose rule found nothing produce the same empty
	// finding list, and the differential harness cannot tell them apart from the outside. This is how
	// it asks.
	listRules := flag.Bool("rules", false, "print the rules this cohere implements for the project's language, one per line, and exit")
	// `-rules` answers what the binary CAN run; this answers what it WILL. The two differ by every
	// rule the config never names, and that gap is invisible from the outside: a rule the config
	// cannot resolve passes its own fixtures and lints nothing, which reads exactly like a rule that
	// found nothing wrong.
	//
	// Reconstructing the answer by hand is the alternative, and it is worse than it looks. Doing it
	// against this config meant walking a nested rules block, then the overrides, then deciding
	// whether a scoped `off` counts, and the hand-derived set was wrong on 29 of 356 rules when
	// finally checked against the real resolver. ESLint has had `--print-config` for this reason.
	listRulesEnabled := flag.Bool("rules-enabled", false,
		"print the rules the lint config resolves for one file (the path given, else index.ts at the project root), with severity, and exit")
	printConfig := flag.Bool("print-config", false,
		"print, as JSON in ESLint's --print-config shape, every registered rule's resolved severity and options for one file (the path given, else index.ts at the project root), and exit")
	// A bare run, --fix and --no-fix format by default, so naming --format there changes nothing. It is
	// accepted rather than refused, so the commands written before the default still mean what they did.
	format := flag.Bool("format", false, "format the files not on record as formatted, or the paths named; a bare run, "+
		"--fix and --no-fix already do, so naming it there changes nothing")
	maxFixPasses := flag.Int("fix-passes", edit.DefaultMaxPasses, "how many times a file may be re-linted while fixes keep landing")
	// Readiness is measured only when asked (#9tgm3dq): unasked, none of cohere:adamic's measure-only rules runs,
	// so a default run pays nothing for it. The set's own findings report whenever a project enables it; this
	// flag governs only the measurement.
	adamicReadiness := flag.Bool("adamic-readiness", false,
		"say what share of the files are Adamic-ready, measured with cohere:adamic at the set's own options, in the footer and in --json")
	showTiming := flag.Bool("timing", false, "report what building the graph cost and the CPU each rule cost, most expensive rule first")
	explainFile := flag.String("explain", "", "report what every rule did on one file, and why it did or did not run, writing nothing")
	// The counts print on every run; this names every rule once under the one coverage fact that
	// describes it. Behind a flag because 170 per-rule notes on a clean run buried the lines that need
	// action, and in front of nobody's habit because the counts that add up stay on the default line.
	showCoverage := flag.Bool("coverage", false, "name every rule once under the coverage fact that describes it, rather than only counting them")
	unusedReport := flag.Bool("unused", false, "report code that was written and never used: unreferenced exports, and statements nothing can reach")
	unusedAll := flag.Bool("unused-all", false, "list the unused findings already marked cohere-keep rather than only counting them (implies --unused)")
	// Opt-in on purpose. The closure's claim is strictly stronger than the flat one and it fails
	// differently: a wrong root mis-reports one file in the flat view and cascades here, going dark
	// across everything that was alive only through it. Keeping both means the two numbers can be
	// read against each other before the bigger one is trusted.
	unusedDeep := flag.Bool("unused-deep", false, "also compute the transitive closure of unused code and group it into islands (implies --unused)")
	showVersion := flag.Bool("version", false, "print the version, what this binary was built from, and the Swift contract it speaks, and exit")
	cacheDump := flag.Bool("cache-dump", false, "print what the cache table for this project holds, and exit")
	// The editor's save: the buffer arrives on stdin and what --fix (and --format) would write for this
	// path leaves on stdout, with nothing written to disk. See stdin.go.
	stdinFilePath := flag.String("stdin-filepath", "",
		"with --fix, read one file's text from stdin and print what --fix (and --format, if named) would write for the file at this path, writing nothing to disk")
	// A profile of a cold run: a profiled run is never one the run cache replays, since only a bare run
	// or `--no-fix` is.
	// The backquoted word is the argument's name in the help, which is how the flag package spells one.
	profilePath := flag.String("profile", "", "write a Go CPU profile of the run to `file`, for go tool pprof")
	// The view. A run prints the files it rewrote, its findings and one footer line; these widen it.
	verbose := flag.Bool("verbose", false,
		"print everything a run can say: each phase, the coverage summary, overrides, skips, notes, memory and the total")
	showPhases := flag.Bool("phases", false, "put where the time went first in the footer's parentheses")
	jsonOutput := flag.Bool("json", false, "print newline-delimited JSON for a program to read, described by schema/CohereOutput.schema.json")
	flag.Parse()
	cacheOff = *noCache
	if *verbose && *jsonOutput {
		return fmt.Errorf("--verbose and --json contradict each other: --verbose prints the human account in full, --json prints JSON for a program")
	}
	adamicReadinessRequested = *adamicReadiness
	activeSummary.AdamicRequested = *adamicReadiness
	activeOutput = outputSettings{Mode: outputHuman, Phases: *showPhases, Style: styleFor(os.Stdout)}
	switch {
	case *verbose:
		activeOutput.Mode = outputVerbose
	case *jsonOutput:
		activeOutput = outputSettings{Mode: outputJSON}
	}
	if *profilePath != "" {
		if err := startProfile(*profilePath); err != nil {
			return err
		}
	}

	// Where the project is, decided once, before anything reads a path.
	//
	// The failure is held rather than returned here, because `-rules`, `-version` and an explicitly
	// named lint config do not need a project at all, and refusing them for standing outside one would
	// be a refusal with no reason behind it. Every path that does need one returns it at the point of
	// use.
	//
	// Decided before the listings rather than after them, because inside a Swift package every one of
	// them is the Swift engine's question: `cohere --rules` there lists the rules that would run there.
	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolving the working directory: %w", err)
	}
	given := map[string]bool{}
	flag.Visit(func(set *flag.Flag) { given[set.Name] = true })
	location, locateError := locateProject(locationRequest{
		WorkingDirectory:        workingDirectory,
		Directory:               *directory,
		ConfigFileName:          *configFileName,
		ConfigFileNameGiven:     given["tsconfig"],
		LintConfigFileName:      *lintConfigFileName,
		LintConfigFileNameGiven: given["lint-config"],
		Engine:                  projectEngine(os.Getenv(projectEngineVariable)),
	})

	startRunCacheClock(location, locateError)

	// A root holding more than one project, TypeScript and Swift side by side or nested, checks each with
	// its own engine and reports them as one (#f9nftxz). One project is checked here, as it always was.
	if discoveryApplies(given, flag.Args()) {
		namedLintConfig := ""
		if given["lint-config"] {
			namedLintConfig = absoluteFrom(workingDirectory, *lintConfigFileName)
		}
		exitCode, checked, err := checkDiscoveredProjects(location, locateError, workingDirectory, namedLintConfig)
		if err != nil {
			return err
		}
		if checked {
			exitProcess(exitCode)
		}
		// A test instrument: a file created in the moment after discovery listed the tree and before anything else
		// reads it, holding COHERE_TEST_CREATE_AFTER_DISCOVERY_TEXT when that is set. See
		// TestAFileCreatedAfterDiscoveryIsNeverReplayedOver and TestAFileCreatedAfterDiscoveryIsFormattedNextRun.
		if created := os.Getenv("COHERE_TEST_CREATE_AFTER_DISCOVERY"); created != "" {
			text := "export const late: number = \"x\";\n"
			if given, set := os.LookupEnv("COHERE_TEST_CREATE_AFTER_DISCOVERY_TEXT"); set {
				text = given
			}
			if err := os.WriteFile(created, []byte(text), 0o644); err != nil {
				return err
			}
		}
	}

	if locateError == nil && location.Engine == engineSwift {
		exitCode, err := runSwiftEngine(location, given, flag.Args())
		if err != nil {
			return err
		}
		if exitCode != 0 {
			exitProcess(exitCode)
		}
		return nil
	}

	if *listRules {
		// Sorted so two binaries can be diffed directly. The registry's own order is the order rules
		// were added, which is meaningful to a reader and useless to `diff`.
		names := make([]string, 0, len(registry.All()))
		for _, registered := range registry.All() {
			names = append(names, registered.Name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Println(name)
		}
		// The provenance travels with the list rather than waiting behind `-version`.
		//
		// A development build implements whatever was on disk when it was compiled, which in a shared
		// worktree includes another author's uncommitted work. Three of us read a count off this
		// surface on one night and reported it as the repository's, and `-version` had been saying so
		// the whole time. A disclosure that has to be sought is a disclosure that will be skipped,
		// especially when the number it qualifies looks plausible, so the qualification has to travel
		// with the number a reader carries away rather than sitting one command to the side.
		//
		// It goes to stderr so the rule list on stdout stays diffable between two binaries.
		if provenance := release.Current(); provenance.IsDevelopment() {
			fmt.Fprintln(os.Stderr, rulesProvenanceNote(provenance, len(names)))
		}
		return nil
	}

	if *listRulesEnabled || *printConfig {
		lintConfigPath := location.LintConfigFileName
		if locateError != nil {
			if !given["lint-config"] {
				return locateError
			}
			lintConfigPath = absoluteFrom(absoluteFrom(workingDirectory, *directory), *lintConfigFileName)
		}
		if locateError == nil && usesHouseSets(location) {
			return errHouseSetsPerFile
		}
		lintConfig, err := configuration.LoadFor(lintConfigPath, registeredRuleNames())
		if err != nil {
			return fmt.Errorf("reading %s: %w", lintConfigPath, err)
		}

		// Resolution is per file, because an override can turn a rule off for one path and leave it
		// on everywhere else, so there is no single answer for the whole tree. The probe path names
		// which file the answer is about, and it defaults to a plain TypeScript source at the root
		// rather than to nothing: a caller who forgets to pass one gets the ordinary case rather
		// than an error, and a caller who wants the generated-file answer asks for it by name.
		//
		// A typed probe is resolved where it was typed. Left relative, the config reads it as relative
		// to its own directory, so asking about `Thing.ts` from `modules/tasks` answered for a file at
		// the root that does not exist.
		probePath := "index.ts"
		if arguments := flag.Args(); len(arguments) > 0 {
			probePath = absoluteFrom(absoluteFrom(workingDirectory, *directory), arguments[0])
		}
		resolved := lintConfig.Resolve(probePath)
		if resolved.Ignored {
			fmt.Fprintf(os.Stderr, "note: %s is excluded by ignore pattern %q, so no rule applies to it\n",
				probePath, resolved.IgnoredBy)
			return nil
		}
		if *printConfig {
			return writeResolvedConfig(os.Stdout, resolved)
		}

		// Only rules the binary actually implements. A config key naming a rule this build does not
		// have resolves to a severity and still runs nothing, and printing it here would report a
		// rule as enabled that cannot fire. That gap is real rather than hypothetical: this config
		// carries several keys spelled for an older name, and they are exactly the rules whose
		// standing decision silently stopped applying.
		implemented := make(map[string]bool, len(registry.All()))
		for _, registered := range registry.All() {
			implemented[registered.Name] = true
		}

		type enabledRule struct {
			name     string
			severity string
		}
		var enabled []enabledRule
		unimplemented := 0
		for name, setting := range resolved.Rules {
			if setting.Severity == configuration.SeverityOff {
				continue
			}
			if !implemented[name] {
				unimplemented++
				continue
			}
			enabled = append(enabled, enabledRule{name: name, severity: setting.Severity.String()})
		}
		sort.Slice(enabled, func(first, second int) bool { return enabled[first].name < enabled[second].name })
		for _, rule := range enabled {
			fmt.Printf("%s\t%s\n", rule.name, rule.severity)
		}

		fmt.Fprintf(os.Stderr, "note: %d rules enabled for %s, of %d implemented by this binary\n",
			len(enabled), probePath, len(implemented))
		if unimplemented > 0 {
			fmt.Fprintf(os.Stderr,
				"note: %d config keys name a rule this binary does not implement, so their setting applies to nothing\n",
				unimplemented)
		}
		return nil
	}

	if *showVersion {
		// The full provenance rather than a bare number, because the number alone does not identify
		// what ran: the pinned typescript-go commit is most of the code in this binary and moves
		// independently of the version. A bug report that names all of it is reproducible.
		fmt.Println(versionReport())
		return nil
	}

	if *cacheDump {
		if locateError != nil {
			return locateError
		}
		return dumpCacheTable(location)
	}

	// A bare `cohere` runs the whole pipeline. The phase flags isolate one phase for someone
	// debugging it; they are not the default path, and nothing about running the tool with no
	// arguments should be a partial check.
	anyPhaseNamed := *typesOnly || *lintOnly || *fixOnly || *formatOnly
	runFix := *fixOnly || *formatOnly || !anyPhaseNamed
	runTypes := *typesOnly || !anyPhaseNamed
	runLint := *lintOnly || !anyPhaseNamed

	// `--unused` is the one phase a bare `cohere` does not run, and it is deliberately absent from
	// `anyPhaseNamed` above rather than folded into it.
	//
	// Two separate things follow from that, and both are the point. It does not run unless named, so
	// the gate people run constantly stays the correctness gate and nothing else. And naming it does
	// not narrow the run the way `--lint` does — `cohere --unused` still fixes, type-checks, and
	// lints, because unused leans on the type graph being sound and a narrowing flag would encourage
	// running it against a tree whose symbols do not resolve.
	//
	// The reason it is opt-in rather than merely quiet: this is a report someone asks for, not part
	// of the correctness gate. An export kept for an external consumer is unused and correct, so
	// there is no answer here a build can enforce, and a phase that cannot fail honestly should not
	// be in the path of a phase that can.
	runUnused := *unusedReport || *unusedAll || *unusedDeep

	// Carried to every bail site so an opt-in phase that was genuinely cut off reads differently
	// from one nobody asked for. Without it a bail prints "unused did not run (types bailed)" on
	// every ordinary run, which manufactures a gap out of a run where nothing was withheld.
	unusedRequest := requestedPhases{phaseUnused: runUnused}

	// `--explain` explains the file as it stands, so it writes nothing: it is a `--no-fix` run with an
	// explanation after it. It used to run the writing fix phase, so `cohere --explain SecretRow.tsx` in
	// ahra rewrote two other files, rebuilt the graph, and explained a tree that was no longer the one
	// asked about (#sm79kfv). Naming it with `--fix`, which only writes, is refused by name.
	if *explainFile != "" {
		if *fixOnly {
			return fmt.Errorf("--explain and --fix contradict each other: --explain reads the file as it stands and writes nothing, --fix only writes")
		}
		*noFix = true
	}

	// `--no-format` leaves formatting out, so a flag that asks for formatting beside it is a contradiction.
	if *noFormat {
		for _, other := range []struct {
			named bool
			flag  string
		}{{*format, "--format"}, {*formatAll, "--format-all"}, {*formatOnly, "--format-only"}} {
			if other.named {
				return fmt.Errorf("--no-format and %s contradict each other: --no-format leaves formatting out, and %s asks for it", other.flag, other.flag)
			}
		}
	}

	// `--format-only` is formatting and nothing else, so naming it beside a flag that asks for other work
	// is a contradiction, refused by name like `--fix --no-fix` rather than resolved one way silently.
	if *formatOnly {
		for _, other := range []struct {
			named bool
			flag  string
			asks  string
		}{
			{*fixOnly, "--fix", "applies lint fixes"},
			{*typesOnly, "--types", "runs the type check"},
			{*lintOnly, "--lint", "runs the lint rules"},
			{runUnused, "--unused", "runs the unused-code report"},
			{*explainFile != "", "--explain", "explains what the rules did"},
		} {
			if other.named {
				return fmt.Errorf("--format-only and %s contradict each other: --format-only runs formatting alone, and %s %s", other.flag, other.flag, other.asks)
			}
		}
	}

	// `--no-fix` mutates nothing, which is what continuous integration needs and what anyone asking
	// "what would this change" needs. The fix phase still runs under it, in memory: every file a
	// writing run would rewrite is reported as a finding, so an unformatted file cannot pass a clean
	// run. Naming it with `--fix` is refused below.
	mutate := runFix && !*noFix
	if *noFix && *fixOnly {
		// Naming both is a contradiction the user should hear about rather than have silently
		// resolved, since one of the two was certainly not what they meant.
		return fmt.Errorf("--fix and --no-fix contradict each other: --fix runs only the mutating phase, --no-fix mutates nothing")
	}

	ctx := context.Background()

	// Everything from here checks a project, so standing outside one stops the run, naming where it
	// looked.
	if locateError != nil {
		return locateError
	}

	if given["stdin-filepath"] {
		// One file, from stdin, to stdout: no other phase runs and no report is printed, because stdout
		// is the file's text. --fix is required rather than implied, so the command line says what the
		// output means.
		if !*fixOnly || *noFix || *typesOnly || *lintOnly || len(flag.Args()) > 0 {
			return errStdinNeedsFix
		}
		return runStdin(ctx, stdinRequest{
			Location:         location,
			WorkingDirectory: workingDirectory,
			FilePath:         *stdinFilePath,
			NoFormat:         *noFormat,
			MaxPasses:        *maxFixPasses,
			SingleThreaded:   *singleThreaded,
		}, os.Stdin, os.Stdout)
	}

	// The repository's own way of printing, from its settings, under whatever the command line named.
	block, err := readOutputBlock(location.LintConfigFileName)
	if err != nil {
		return err
	}
	if block.Phases != nil && !given["phases"] {
		activeOutput.Phases = *block.Phases
	}

	// The run cache: a bare run whose every input is unchanged replays its recorded report here and
	// exits, before the graph is built. Otherwise this starts recording. See run_cache.go.
	runCacheInputs := beginRunCache(location)

	// The formatter, built here rather than in the fix phase, so a bare run can start the format walk now,
	// beside the graph build, instead of after it: the walk reads only the disk, and on ahra it was about
	// 100ms between the graph and the fix phase on a cold run (#679s763). A formatter that cannot load is
	// reported where it always was. See startFormatWalkAhead.
	var formatter formatEngine
	var formatterError error
	if runFix {
		formatter, formatterError = configuredFormatter(!*noFormat && (*explainFile == "" || *format || *formatAll))
	}
	// And the repository the run may write: the project's, or, for a run started inside a library that is a
	// repository of its own (libraries/structure in ahra), that library's. Writes stay inside one
	// repository (@system_cohere, 2026-10-03), so the run there formats and fixes the library, and its
	// check still reads the whole program. See nested_repository.go. Computed once, here, since the format
	// walk ahead and the early format pass read the tree it names before the fix phase does.
	repositoryRoot := writeRepositoryRoot(location.ArgumentBase, location.Root)
	if formatter != nil && len(flag.Args()) == 0 {
		// Walked from discovery's listings where it has them (#g3046x5), when discovery ran.
		var listing gitignore.Listing
		if discoveredListings != nil {
			listing = discoveredListings.Listing
		}
		startFormatWalkAhead(formatter, repositoryRoot, listing)
	}

	// A bare run's format pass begins now, on the cores the graph build leaves idle, rather than with the walk
	// (#g3046x5). Only the default scope, the one a bare run takes below. See earlyFormat.
	var early *earlyFormat
	if formatter != nil && len(flag.Args()) == 0 && runFix && !*formatAll && !*formatOnly && *explainFile == "" &&
		parallelFormatWorkers() != 0 {
		earlyRecord := formatRecordOff("the cache is off (--no-cache)")
		if !cacheOff {
			earlyRecord = loadFormatRecord(repositoryRoot)
		}
		early = startEarlyFormat(formatter, earlyRecord, repositoryRoot, *maxFixPasses)
	}

	// What this run leaves to nearer projects, when it is one project of several (see ownership.go).
	yield, err := activeProjectYield()
	if err != nil {
		return err
	}

	// Under --timing the build says what it was made of, not only how long it took (#cazsft3).
	var graphTiming *program.GraphTiming
	if *showTiming {
		graphTiming = &program.GraphTiming{}
	}
	// `--format-only` builds no graph (#m0dktbn). Formatting reads the disk, never the program: its scope is
	// the format walk, it asks no rule for a fix, and no type check or lint runs, so the build was about
	// 250ms of every commit gate spent on a program nothing read. Without it graph stays nil and the
	// project files empty, and every use below is on a path `--format-only` does not take.
	var graph *program.Graph
	var buildDuration, contentPackOpened time.Duration
	// Read beside the build rather than after it, when the config needs nothing from the program. See
	// lintConfigAhead.
	var lintAhead *lintConfigAhead
	if *formatOnly {
		// The run cache records a run only once its inputs are declared. The format walk declares its own
		// (declareFormatWalk), settings chain included, so there is nothing else to name: the run reads no
		// tsconfig and no lint config.
		declareRunCacheInputs()
	} else {
		if runFix || runLint {
			lintAhead = startLintConfigAhead(location)
		}
		buildStart := time.Now()
		contentPack := openContentPack(location.Root)
		contentPackOpened = time.Since(buildStart)
		built, err := program.Build(program.Options{
			ConfigFileName:   location.ConfigFileName,
			CurrentDirectory: location.Root,
			SingleThreaded:   *singleThreaded,
			Inputs:           runCacheInputs,
			ContentPack:      contentPack,
			CheckedStats:     runCacheCheckStats,
			Timing:           graphTiming,
			Yielded:          yield.yieldedFiles(),
			Listings:         discoveredListings,
		})
		if err != nil {
			// A program that fails to build is a loud failure and never an empty result. An empty file list
			// is indistinguishable from a clean tree, and that confusion is what let the gate this replaces
			// print green over zero files for days.
			return fmt.Errorf("building the type graph: %w", err)
		}
		graph = built
		buildDuration = time.Since(buildStart)
		if early != nil {
			early.programs.offer(graph)
		}
		// A test instrument: every file walked on a checker other than its own. See program.walkQueue.
		graph.WalkOnForeignCheckers = os.Getenv("COHERE_TEST_FOREIGN_CHECKERS") != ""

		// The checkers are created now, alongside the work before the walk, rather than by the walk's first
		// lookup: 74 to 95ms on ahra that otherwise sits on the path to the first worker (#zqsdzbq, lever B).
		// Upstream's pool creates them once, whoever asks first, so the walk and the types check just find
		// them. Any phase that runs reads them.
		if projectFiles := graph.ProjectFiles(); len(projectFiles) > 0 && (runFix || runTypes || runLint) {
			go func() {
				_, release := graph.Program.GetTypeCheckerForFile(ctx, projectFiles[0])
				release()
			}()
		}

		// The build saw every file the compiler read. The lint config is read by the command, not the
		// compiler, so it is named here, with every file it extends: a base edited alone changes what runs.
		declareLintConfigInputs(location.LintConfigFileName)
		declareRunCacheInputs(location.ConfigFileName)
	}

	var projectFiles []*ast.SourceFile
	if graph != nil {
		projectFiles = graph.ProjectFiles()
	}
	wholeProgramCount := len(projectFiles)

	// A named path narrows every phase below, because this slice is what they are handed. The type
	// graph is untouched: it still holds the whole program, so a rule that reads a declaration out
	// of a file nobody named still finds it. Only the walk list shrinks.
	//
	// Before this, a positional path was parsed and discarded. `cohere --lint OneFile.ts` took
	// 3.042s and printed 5,201 findings; the whole tree took 3.003s and printed 5,201. Same cost,
	// same output, and nothing in either run said the argument had been ignored.
	lintScope := formatScope{Everything: true}

	// What this run may write, which is a different set from what it checks and is decided once,
	// here, from what the caller stated. Nothing below widens it.
	//
	// The checked set grows on purpose: a named file's importers are checked because its exported
	// types reach them, and past the closure limit the whole tree is. Both are about seeing more.
	// Writing used to ride along with them, so `cohere --fix app` in www-phi-health fell back to the
	// whole tree and rewrote 15 files in libraries/structure and 2 in its nested nexus, none of them
	// named. A caller who names paths has said which files the run is about, and a fixable finding
	// outside them is reported, not repaired. With nothing named the whole project is the caller's.
	writeScope := formatScope{Everything: true}

	if len(flag.Args()) > 0 && !*listRules && !*listRulesEnabled {
		scope, err := namedPathsScope(location.ArgumentBase, location.Root, flag.Args())
		if err != nil {
			return err
		}
		writeScope = scope
		if graph != nil {
			lintScope, projectFiles = narrowToClosure(graph, scope, projectFiles)
		}
	}

	switch {
	case graph == nil:
		fmt.Fprintln(accountOutput(invocationOutput(os.Stdout)), "graph not built: --format-only reads the disk, not the program")
	case lintScope.Everything:
		fmt.Fprintf(accountOutput(invocationOutput(os.Stdout)),
			"graph built in %s — %d files in the program, %d of them ours\n",
			round(buildDuration), len(graph.SourceFiles()), wholeProgramCount,
		)
	default:
		// Both numbers, because a reader who sees only the narrowed count cannot tell a scoped run
		// from a tree that shrank, and those want opposite reactions.
		// The scope's own count is the paths it enumerated, which is not the number checked: a named
		// directory holds markdown and data files the program never contained. Reporting the walked
		// count beside the enumerated one made `58 in scope (2685 named paths)`, two true numbers
		// that read as a contradiction. The enumerated count is dropped and the description says
		// what was asked for instead.
		// The dependent count is named rather than folded into the total, because a reader who asked
		// for one file and sees eleven checked should be told why without having to know that this
		// binary walks imports at all.
		if lintScope.DependentCount > 0 {
			fmt.Fprintf(accountOutput(invocationOutput(os.Stdout)),
				"graph built in %s — %d files in the program, %d of them ours, %d in scope (%s plus %d that import it)\n",
				round(buildDuration), len(graph.SourceFiles()), wholeProgramCount,
				len(projectFiles), lintScope.RequestDescription, lintScope.DependentCount,
			)
		} else {
			fmt.Fprintf(accountOutput(invocationOutput(os.Stdout)),
				"graph built in %s — %d files in the program, %d of them ours, %d in scope (%s)\n",
				round(buildDuration), len(graph.SourceFiles()), wholeProgramCount,
				len(projectFiles), lintScope.RequestDescription,
			)
		}
	}

	// A file the tsconfig includes and a nearer one owns is that one's run's to report, so this one's count
	// says where the rest went (see ownership.go).
	if graph != nil && graph.YieldedFiles > 0 {
		fmt.Fprintf(accountOutput(invocationOutput(os.Stdout)),
			"left %d files the tsconfig includes to the nearer tsconfigs that own them, whose runs check them\n", graph.YieldedFiles)
	}

	if graphTiming != nil && graph != nil {
		printGraphTiming(os.Stdout, graphTiming, buildDuration, contentPackOpened, graph.Workers())
	}

	findings := 0
	// What the footer says, filled in as the phases finish. Lint replaces the counts with its own; a run
	// lint does not reach still accounts for every file in scope.
	activeSummary.FilesInScope, activeSummary.FilesChecked = len(projectFiles), len(projectFiles)
	activeSummary.Gaps.NothingToCheck = graph != nil && len(projectFiles) == 0
	activeSummary.Gaps.FormattingNotChecked = !runFix
	report := &pipelineReport{
		graph:          buildDuration,
		processStart:   processStart,
		filesInScope:   len(projectFiles),
		filesInProgram: wholeProgramCount,
		graphSkipped:   graph == nil,
		rootNote:       location.rootNote(),
		cacheOff:       cacheOff,
	}

	// Phase 2: fix and format. Mutation runs before anything reports, so every phase downstream sees
	// the repaired tree rather than findings a fixer would have silently repaired.
	//
	// The lint config is loaded here because fixing needs to know which rules apply to which files,
	// and loading it once serves both this phase and lint below.
	var lintConfig *configuration.Config

	// The fix phase's walk, kept when it is still valid for the lint phase to reuse.
	//
	// Nil means lint must walk for itself: either the fix phase did not run, or it rewrote a file and
	// the graph was rebuilt underneath these diagnostics. Both cases are the same instruction, which
	// is why one nil check covers them.
	var reusableWalk *program.Result

	// `--format-only` proposes no fixes, so it reads no rules and its verdict cannot hang on their config.
	if (runFix && !*formatOnly) || runLint {
		loaded, err := configureLint(graph, location, lintAhead)
		if err != nil {
			return err
		}
		lintConfig = loaded
		writeSetsLine(accountOutput(os.Stdout), lintConfig)
		// The run cache's second layer, attached once the config is known and before anything walks.
		attachFindingsCache(graph, location)
	}
	// A run the run cache does not record still replays and records the types section. See typesRecord.
	if runTypes {
		attachTypesCache(graph, location)
	}
	if graph != nil {
		report.contentKeyedShapes = graph.ContentKeyedShapes
	}

	// The types phase's checking starts here, alongside the fix phase's walk, rather than after it. See
	// startTypeCheck. Its result is used only if the walk leaves this graph in place.
	//
	// A run with nothing to replay and no build info to read checks the whole program, and that check is done
	// inside the fix phase's walk instead, file by file just before each file's rules (#679s763): beside the
	// walk, the two fought over every checker. See program.FusedCheck.
	var earlyTypeCheck *typeCheck
	switch {
	case runTypes && runFix && cacheOff && activeTypesReuse() == nil && !*formatOnly && len(projectFiles) > 0:
		graph.FusedCheck = program.NewFusedCheck()
	case runTypes && runFix:
		earlyTypeCheck = startTypeCheck(ctx, graph, !cacheOff)
	}

	switch {
	case !runFix:
		report.record(phaseFix, outcomeSkipped, 0, "not requested")
	default:
		// Under `--no-fix` everything below runs exactly as it does for a writing run, scope and
		// formatter included, and only the write is withheld: see applyProposedFixes. It used to skip
		// the phase, and a tree with an unformatted file printed `lint: 0 findings` over it on every
		// clean run, with `fix skipped (--no-fix)` the only word on the phase that would have seen it.
		//
		// Built before the scope rather than after it, so a formatter that cannot load its bundles stops
		// the run here with a reason rather than degrading into the nil that means nobody asked for one,
		// and so the scope can know whether there is anything to format at all.
		//
		// `--format-all` asks for formatting by naming its scope, so it turns the formatter on. Alone it
		// used to configure none: `cohere --format-all` formatted nothing, and `cohere --no-fix
		// --format-all` checked nothing, each printing a clean run.
		//
		// A run formats by default (Kirk's ruling, 2026-10-04: "running cohere should type check lint fix
		// format all in one call"), unless --no-format leaves it out or the run is --explain, which explains
		// the rules on one file and writes nothing.
		if formatterError != nil {
			return formatterError
		}
		activeSummary.Gaps.FormattingNotChecked = formatter == nil
		var formatLeftOut string
		if formatter == nil && *noFormat {
			formatLeftOut = "formatting left out (--no-format)"
		}

		// The format record: which bytes cohere has already seen formatted. Every formatting run adds to
		// it, and the default scope is read from it. See format_record.go.
		var record *formatRecord
		switch {
		case early != nil:
			record = early.record
		case formatter != nil && cacheOff:
			record = formatRecordOff("the cache is off (--no-cache)")
		case formatter != nil:
			// The record lives with the repository the run writes, because saving it prunes every entry
			// outside the universe it was saved with: a run inside a library keeping ahra's record would
			// drop all of ahra's entries.
			record = loadFormatRecord(repositoryRoot)
		}
		// Every file the default scope was drawn from, so the record can drop entries outside it. Nil for a
		// narrower scope, which says nothing about the files it did not look at.
		var recordUniverse []string

		// The format phase gets its own universe, and this is where the two narrowings part.
		//
		// A proposed fix comes from a rule that ran over the program, so its candidate must be in the
		// program. A format candidate comes from the disk, and intersecting it with the type graph is
		// what made css, markdown, json and yaml invisible: a tsconfig enumerates TypeScript by
		// construction. Formatting is the only phase whose subject is not the program.
		var scope formatScope
		// Drawn on the main goroutine, but not always before the walk. A bare run's default scope is the early
		// pass's, which is ready only once that pass has enumerated the whole format universe, and on a cold run
		// that came 36 to 51ms after the graph, with the walk waiting on it at 6 to 9 busy cores (#g3046x5). The
		// walk needs neither the scope nor that enumeration, so with an early pass the fix phase draws the scope
		// after its walk, through formatPlan. Every other scope is drawn here, as it always was.
		scopeDrawn := false
		drawScope := func() {
			if scopeDrawn {
				return
			}
			scopeDrawn = true
			switch {
			case formatter == nil && writeScope.Everything:
				// Nothing will be formatted, so there is nothing to find. This used to ask git what changed on
				// every bare run, which was about half of a cached replay, and every file it found then reported
				// as not formatted because nobody asked: a cost and a count that were both about nothing.
				description := "nothing, since formatting was not requested"
				if formatLeftOut != "" {
					description = "nothing: " + formatLeftOut
				}
				scope = formatScope{index: map[string]struct{}{}, Description: description}

			case formatter == nil:
				// No formatter, and paths were named. The scope keeps its type-graph narrowing so the fix
				// phase's own reporting is unchanged.
				//
				// Narrowed against the whole program rather than against the named scope, because the format
				// phase has its own universe: reporting the intersection with the scope instead made `7 changed
				// files, 0 of them in the program` on a tree where three of them were.
				wholeProgram := graph.ProjectFiles()
				inProgram := make(map[string]struct{}, len(wholeProgram))
				for _, sourceFile := range wholeProgram {
					// Cleaned to the form the named paths have, which on Windows is the `\` the compiler's
					// names do not use.
					inProgram[filepath.Clean(sourceFile.FileName().AsString())] = struct{}{}
				}
				scope = writeScope.narrowTo(inProgram)

			case !writeScope.Everything && len(writeScope.FileNames) == 0:
				// Named paths that hold no file, so the walk cannot affect the outcome. The scope already names
				// where it looked, which is the honest thing to print here.
				scope = writeScope

			case !writeScope.Everything || *formatAll:
				// A caller who named paths has already said which files this run is about, and formatting
				// files they did not name would be a surprise in the one mode where they were explicit. It reads
				// the write scope and not the check scope: the check scope becomes the whole tree past the
				// closure limit, and keying on it sent a named run down to the default below, formatting every
				// changed file in every submodule. It also comes before `--format-all`, because with paths
				// stated, all of them is all of the stated paths.
				//
				// `--format-all` alone is the whole walk, asked for by name.
				//
				// The enumeration walks the tree through ignore layers, which are inputs the run cache does not
				// observe. Declining here costs a miss in this configuration and never a stale hit.
				scope = wholeTreeScope()
				if !writeScope.Everything {
					scope = writeScope
				}
				declineRunCache("a formatter enumerated the tree")
				enumeration, enumerateError := enumerateFormatTree(formatter, repositoryRoot)
				if enumerateError != nil {
					// A failed walk withholds formatting and says why, rather than falling back to a universe
					// that would format the wrong set. Fixing still runs.
					scope = formatScope{Description: fmt.Sprintf("nothing (could not enumerate the tree: %v)", enumerateError), failure: enumerateError}
				} else {
					scope = scope.narrowToEnumeration(enumeration)
				}

			default:
				// What changed since cohere last looked: every file the formatter handles, here and in each
				// declared submodule, whose bytes are not on record as formatted. Measured on ahra's 3,084
				// formattable files, the native printers format the whole tree in about 10 seconds, which the
				// first run with no record pays once; every run after formats what was edited.
				//
				// The walk declares what it read, so a run that only checks can be replayed (#13a63n3). The scope it
				// draws also reads the format record, which changes nothing a replay could print wrongly: an entry is
				// a proof that bytes are already formatted, so the record decides how much is formatted again, never
				// what would change.
				if early != nil {
					scope, recordUniverse = early.formatScopeOf()
				} else {
					scope, recordUniverse = unformattedScope(formatter, record, repositoryRoot)
				}
			}

			// The Adamic `.a` files the program holds join the format scope now that there is a program to ask
			// (#6mhafvb). The walk held every `.a` back, since one could as well be a static library, and the early
			// pass never saw one. A held `.a` git ignores is checked and not formatted, and the run says so. A nested
			// repository's drift check has no program to ask, so it reads none of a library's `.a` files: their
			// drift is the library's own run to find, not this one's.
			if formatter != nil {
				held := adamicHeld(graph)
				if graph == nil && len(scope.adamic) > 0 {
					// The tsconfig is an input this run declared it never reads, so it is not replayed.
					declineRunCache("the tsconfig says which .a files are source")
					included, includedError := adamicIncluded(location.ConfigFileName)
					if includedError != nil {
						// Said rather than swallowed: the `.a` files stay out of the scope, and the reason is the run's.
						fmt.Fprintf(os.Stderr, "⚠ the Adamic .a files were not formatted: reading the tsconfig to see which are source: %v\n", includedError)
					}
					held = included
				}
				var claimedAdamic []string
				scope, claimedAdamic = scope.claimAdamic(held)
				if recordUniverse != nil {
					recordUniverse = slices.Concat(recordUniverse, claimedAdamic)
				}
				activeSummary.Gaps.AdamicIgnored = reportIgnoredAdamic(os.Stderr, repositoryRoot, held)
			}
		}
		if early == nil {
			drawScope()
		}

		// The submodules this repository declares are read, never written: each file a run inside one
		// would rewrite is a finding here, so a library's drift fails the project's check until that
		// library's own run formats it. Only on a whole-tree check with a formatter, since a run that named
		// paths has said which files it is about. It starts here, beside the fix phase, and is joined below.
		var nestedCheck <-chan nestedCheckResult
		if !mutate && formatter != nil && writeScope.Everything {
			nestedCheck = startNestedCheck(formatter, repositoryRoot, *formatAll)
		}

		// `--format-only` asks no rule for a fix: no program files to walk, no rules to walk them with, so the
		// fix phase's candidates are the format scope's and nothing else.
		fixFiles, fixRules := projectFiles, registry.All()
		if *formatOnly {
			fixFiles, fixRules = nil, nil
		}

		fixStart := time.Now()
		formatting := newFormatClock()
		if early != nil {
			formatting = early.formatting
		}
		formatPlan := func() (edit.Transform, []string) {
			drawScope()
			return formatting.timing(scopedTransform(record.observe(formatTransform(formatter), optionsFingerprintOf(formatter)), scope)),
				scope.formatCandidates()
		}
		fixSummary, fixWalk, lineEndings, err := applyProposedFixes(
			ctx, graph, fixFiles, fixRules,
			formatPlan,
			writeScope,
			repositoryRoot,
			*maxFixPasses,
			mutate,
			early,
		)
		fixDuration := time.Since(fixStart)
		// With no formatter the transform only declines each file, which is not formatting.
		if formatter != nil {
			activeSummary.Formatting = formatting.Total()
		}
		if err != nil {
			// The bail condition here is a failure to produce valid output, never a finding. A fixer
			// that cannot write a parseable file means the edit was malformed and everything after it
			// is meaningless, so the pipeline stops and says which phases never ran.
			report.record(phaseFix, outcomeRan, fixDuration, "")
			report.markRemainingNotReachedFor(phaseFix, err.Error(), unusedRequest)
			writeRunEnd(report, os.Stdout)
			return fmt.Errorf("fix: %w", err)
		}
		// True of the tree and actionable, so a replay keeps it, but it says which run produced it.
		fmt.Fprintln(accountOutput(provenanceOutput(os.Stdout)), fixSummary)
		for _, refusal := range fixSummary.Refusals() {
			fmt.Fprintln(accountOutput(provenanceOutput(os.Stdout)), "  "+refusal)
		}
		// The files the fix phase rewrote, above the findings, in the views that list them.
		if mutate {
			activeSummary.Changed = changedFilesFrom(fixSummary.ChangedFiles, location.Root)
			printChangedFiles(os.Stdout, activeSummary.Changed)
		}

		// Not writing the record costs the next run a format of files already formatted, never a skip,
		// so it is a note rather than a failure.
		if err := record.save(recordUniverse); err != nil {
			fmt.Fprintf(accountOutput(os.Stderr), "note: the format record could not be written: %v\n", firstLine(err.Error()))
		}

		// A run that rewrote files must not be replayed. The run cache stats inputs when it records, which
		// is after the rewrite, so the manifest would match the fixed tree and the next run would replay
		// "files rewritten" over a tree it never touched.
		if mutate && fixSummary.FilesChanged > 0 {
			declineRunCache("the fix phase rewrote files")
		}

		// The scope is stated on every run rather than inferred from a file count. Named paths, the whole
		// tree, and what is not on record as formatted are three different answers to "what was
		// formatted", and a reader cannot tell which one they got from a number alone. It describes the run
		// that drew the scope, record included, so a replay says which run that was.
		fmt.Fprintf(accountOutput(provenanceOutput(os.Stdout)), "format scope: %s\n", scope.Description)

		// A run whose verdict is formatting alone counts every file it could not read: an empty scope from
		// a failed walk, or a file the formatter declined because it does not parse, or broke on, would
		// otherwise read as a clean tree. Other runs leave those to the phases that report them.
		if *formatOnly {
			unchecked := formatOnlyUnchecked(fixSummary, scope)
			printUnchecked(os.Stdout, unchecked)
			findings += len(unchecked)
			activeSummary.Findings += len(unchecked)
		}
		if mutate {
			report.record(phaseFix, outcomeRan, fixDuration, "")
		} else {
			// Each file a writing run would change is a finding, and the count joins the verdict: a
			// `--no-fix` run over a tree `--fix` would rewrite is not clean.
			printWouldChange(os.Stdout, fixSummary.ChangedFiles, lineEndings)
			findings += len(fixSummary.ChangedFiles)
			activeSummary.WouldChange += len(fixSummary.ChangedFiles)

			// The nested check started beside the fix phase. What it adds to the phase is only the wait
			// past the fix phase's own end.
			if nestedCheck != nil {
				waitStart := time.Now()
				answer := <-nestedCheck
				fixDuration += time.Since(waitStart)
				if answer.err != nil {
					return fmt.Errorf("reading the nested repositories: %w", answer.err)
				}
				nested := answer.check
				for _, walk := range nested.Walks {
					declareFormatWalk(walk)
				}
				for _, note := range nested.Notes {
					fmt.Fprintln(accountOutput(os.Stderr), note)
				}
				printNestedDrift(os.Stdout, nested)
				findings += len(nested.Drift)
				activeSummary.WouldChange += len(nested.Drift)
				if *formatOnly {
					printUnchecked(os.Stdout, nestedUnchecked(nested))
					findings += len(nested.Unchecked)
					activeSummary.Findings += len(nested.Unchecked)
				}
				fmt.Fprintln(accountOutput(provenanceOutput(os.Stdout)), nestedSummary(nested))
			}
			report.recordChecked(phaseFix, fixDuration, len(fixSummary.ChangedFiles))
		}
		switch {
		case *formatOnly:
			report.narrow(phaseFix, "formatting only, no fixes proposed")
		case formatLeftOut != "":
			report.narrow(phaseFix, formatLeftOut)
		}

		// Files were rewritten, so the graph built from the old bytes no longer describes the tree.
		// Every phase after this must read the new text or it reports findings against source that no
		// longer exists — which is the same stale-read corruption the edit engine refuses internally,
		// one level up.
		if mutate && fixSummary.FilesChanged > 0 && (runTypes || runLint) {
			rebuiltGraph, rebuildDuration, err := rebuildGraph(location.ConfigFileName, location.Root, *singleThreaded, lintConfig, yield.yieldedFiles())
			if err != nil {
				report.markRemainingNotReachedFor(phaseFix, fmt.Sprintf("the graph could not be rebuilt after fixing: %v", err), unusedRequest)
				writeRunEnd(report, os.Stdout)
				return fmt.Errorf("rebuilding the type graph after fixing: %w", err)
			}
			// The early check was of the bytes the fixer just replaced. It is retired and left to finish
			// rather than cancelled: the compiler's whole-program check panics when its context ends part
			// way, which killed the gate itself (#sm79kfv). Its graph records nothing, and its result is
			// dropped with the reference.
			graph.Retire()
			carryCaches(graph, rebuiltGraph)
			graph = rebuiltGraph
			earlyTypeCheck = nil
			// The scope survives the rebuild. Taking every project file here turned a run scoped to
			// one named path, or to what changed, into a whole-tree run the moment a fixer landed:
			// the graph line said `1 in scope` and the types and lint lines then reported every file
			// in the program, with nothing saying the scope had been dropped.
			rescoped, lost := rescopeAfterRebuild(graph.ProjectFiles(), projectFiles, lintScope.Everything)
			for _, fileName := range lost {
				fmt.Fprintf(accountOutput(os.Stderr), "note: %s was in scope and is not in the rebuilt program, so it is not checked\n", fileName)
			}
			// The notes are --verbose's, so the footer says the gap they name.
			if len(lost) > 0 && activeSummary.Gaps.Unread == "" {
				activeSummary.Gaps.Unread = fmt.Sprintf("%d file%s in scope not in the rebuilt program", len(lost), plural(len(lost)))
			}
			projectFiles = rescoped
			fmt.Fprintf(accountOutput(os.Stdout),
				"graph rebuilt in %s — %d files changed, so every later phase reads the new text\n",
				round(rebuildDuration), fixSummary.FilesChanged,
			)
		}

		// The fix phase's walk is reusable by lint only when nothing was rewritten.
		//
		// The condition is the same one the rebuild above turns on, and deliberately so: if any file
		// changed, the graph those diagnostics came from has been replaced and they describe bytes
		// that no longer exist. That is the stale-read the rebuild exists to prevent, and reusing the
		// walk across it would reintroduce it one level up while looking like an optimisation.
		//
		// When nothing changed, the graph is provably the same object the walk ran against, so the
		// result is exactly what a second walk would produce. Measured on the ahra tree: a run costs
		// about 0.9s fixed plus 1.1s per walk, so this removes roughly a third of a default run.
		//
		// With no program file in scope the fix phase did not walk at all, and its empty result is not a
		// walk to reuse: lint walks for itself and refuses the empty set, as it always has.
		//
		// A checked run wrote nothing whatever its count, so its walk is always the graph's own.
		if (!mutate || fixSummary.FilesChanged == 0) && len(projectFiles) > 0 {
			reusableWalk = &fixWalk
		}
	}

	releaseEarlyTypeCheck()

	// The files the type phase reported in, which no readiness can call ready. See readiness.go.
	typeErrorFiles := map[string]bool{}

	// Phase 3: types. This is the phase that bails alone and loudly.
	if !runTypes {
		report.record(phaseTypes, outcomeSkipped, 0, "not requested")
	} else {
		typesStart := time.Now()
		// `--no-fix` promises not to write a byte, and the incremental build info is a byte. It was
		// rewritten on every run regardless, which made the flag's own help text false: measured, the
		// mtime of ahra's tsconfig.tsbuildinfo moved across a `--no-fix` run.
		check := earlyTypeCheck
		if check == nil {
			check = startTypeCheck(ctx, graph, !cacheOff)
		}
		typeDiagnostics := finishTypeCheck(ctx, check, projectFiles, !*noFix)
		// The check's own time plus what finishing it took, rather than the wait here, which is nothing when the
		// check ran alongside the walk. The phase line can then add up to more than the run, and says by how much.
		finishing := check.finished
		if typesStart.After(finishing) {
			finishing = typesStart
		}
		typesDuration := check.elapsed + time.Since(finishing)

		for _, diagnostic := range typeDiagnostics {
			printCompilerDiagnostic(diagnostic)
			if file := diagnostic.File(); file != nil && !graph.IsOptionsDiagnostic(diagnostic) {
				typeErrorFiles[file.FileName().AsString()] = true
			}
		}
		findings += len(typeDiagnostics)
		activeSummary.TypeErrors += len(typeDiagnostics)

		// Counted among the files reported on: the check covers the whole program whatever the scope.
		replayedClause := ""
		if check.replayed > 0 {
			if inScope := check.reuse.ReplayedAmong(projectFiles); inScope > 0 {
				replayedClause = fmt.Sprintf("; %d of %d files' semantic diagnostics replayed from cache", inScope, len(projectFiles))
			}
		}
		fmt.Fprintf(accountOutput(invocationOutput(os.Stdout)),
			"types: %d diagnostics over %d files in %s%s\n",
			len(typeDiagnostics), len(projectFiles), round(typesDuration), replayedClause,
		)
		report.record(phaseTypes, outcomeRan, typesDuration, "")

		// Types gate lint. Rule findings against code whose semantics are wrong are noise the reader
		// has to re-read after fixing the real problem, so one type error alone at the top beats one
		// type error buried under a hundred style findings in a file that does not compile.
		//
		// Bailing here is only honest because the phase line says lint did not run. Without it, a
		// bailed run and a clean lint print the same absence of findings.
		//
		// Only source diagnostics stop it. One about the compiler options, `baseUrl` removed in TypeScript
		// 7 say, is still counted in the verdict above, but it does not make the semantics wrong; where it
		// does, the source fails to type-check and that stops lint. See program.IsOptionsDiagnostic.
		sourceDiagnostics := 0
		for _, diagnostic := range typeDiagnostics {
			if !graph.IsOptionsDiagnostic(diagnostic) {
				sourceDiagnostics++
			}
		}
		if sourceDiagnostics > 0 && runLint {
			report.markRemainingNotReachedFor(
				phaseTypes,
				fmt.Sprintf("%d type diagnostics — lint findings against wrong semantics are noise", sourceDiagnostics),
				unusedRequest,
			)
			// Lint will not report, but the fix phase walked the same rules, and its skips are the ones lint
			// would have named. A skip the types phase covers is a gap here, since that phase stopped the run.
			if reusableWalk != nil {
				activeSummary.Skips = rulesSkippingEveryFile(sumRuleNotes(reusableWalk.Notes), reusableWalk.Coverage.RulesOffered)
			}
			if adamicReadinessRequested {
				activeSummary.Adamic = notMeasured("types bailed")
			}
			writeRunEnd(report, os.Stdout)
			finishRunCache(1)
		}
	}

	if !runLint {
		report.record(phaseLint, outcomeSkipped, 0, "not requested")
	}

	if runLint {
		rules := registry.All()

		// The config was loaded above, before the fix phase, because fixing needs to know which rules
		// apply to which files. Only the timing switch is per-phase.
		graph.CollectTimings = *showTiming

		lintStart := time.Now()

		// Reuse the fix phase's walk when it is still valid, rather than walking the same graph with
		// the same rules a second time.
		//
		// Two walks over one graph was the largest single cost in a default run: measured on the ahra
		// tree, roughly 0.9s fixed plus 1.1s per walk, and the fix phase's own comment described its
		// walk as cheap on the assumption that a reporting walk had already happened. It had not.
		// Fix is phase 2 and this runs after phase 3, so the fix walk was the first and this was the
		// repeat.
		//
		// Validity is decided at the fix phase, not here, because only it knows whether a file was
		// rewritten. See where reusableWalk is set.
		//
		// A run with --timing still walks here, because per-rule timings are collected during the
		// walk and the fix phase ran before CollectTimings was set. Reusing that walk would print an
		// empty timing table, which is a report that looks like a measurement.
		var result program.Result
		reusedWalk := reusableWalk != nil && !*showTiming
		if reusedWalk {
			result = *reusableWalk
		} else {
			// Named paths that reached no program file are refused with the reason rather than with
			// the walk's bare "the file set is empty", which could not say that the file was left out
			// by the tsconfig or ignored by git. See explainNamedPathsOutsideProgram.
			if len(projectFiles) == 0 && !lintScope.Everything && len(flag.Args()) > 0 {
				return errNamedPathsOutsideProgram(location, flag.Args())
			}
			walked, err := graph.Walk(ctx, projectFiles, rules)
			if err != nil {
				return fmt.Errorf("running rules: %w", err)
			}
			result = walked
			reportForeignCheckers(graph, walked, len(projectFiles))
		}
		lintDuration := time.Since(lintStart)

		// How the walk was paid for, rather than a duration that would read as this phase's cost.
		//
		// A reused walk timed here measures the pointer copy, so it prints as `in 0s`: a coverage line
		// stating that 212 rules visited 2 million nodes for free. The counts are real and the
		// duration is the lie, which is the harder kind to spot because everything around it checks
		// out. Naming the phase that did the walking keeps the line checkable.
		lintWalkCost := fmt.Sprintf("in %s", round(lintDuration))
		if reusedWalk {
			lintWalkCost = "walked by the fix phase (nothing was rewritten, so its findings still hold)"
		}

		findings += len(result.Diagnostics)
		if adamicReadinessRequested {
			activeSummary.Adamic = readinessOf(graph, result, runTypes, typeErrorFiles)
		}

		// The findings, the lint line, and the coverage block, which prints unconditionally. See
		// writeLintReport.
		writeLintReport(os.Stdout, lintReport{
			Result:     result,
			Rules:      rules,
			LintConfig: lintConfig,
			WalkCost:   lintWalkCost,
			Details:    *showCoverage,
		})

		// Every phase ran and a file still went unchecked. The phase line cannot express that, so the
		// closing warning names it rather than letting a run that lost a file to a panic end on a clean
		// bill of health. The Swift renderer says the same sentence for the same fact.
		switch {
		case len(result.Coverage.FilesCrashed) > 0:
			report.incompleteBeyondPhases = namedGapsSentence
		case len(result.Coverage.RulesCrashed) > 0:
			report.incompleteBeyondPhases = namedRuleGapsSentence
			activeSummary.Gaps.Unread = counted(len(result.Coverage.RulesCrashed), "rule crash left a file's verdict missing", "rule crashes left files' verdicts missing")
		}

		if *showTiming {
			printTimings(os.Stdout, result.Timings, lintDuration)
		}

		if *explainFile != "" {
			// Explained after the run rather than instead of it, so the reader sees the whole-tree
			// verdict and the single-file account together. Asking why one file behaved the way it
			// did is usually a question about a run that already happened.
			//
			// The typed path is tried where it was typed first, and only then by suffix. From a
			// subdirectory the suffix alone can name a same-named file in another module, and the
			// explanation would be about a file the caller was not asking after.
			subject := findExplainSubject(projectFiles, absoluteFrom(location.ArgumentBase, *explainFile))
			if subject == nil {
				subject = findExplainSubject(projectFiles, *explainFile)
			}
			if subject == nil {
				// Naming the file that was not found rather than explaining nothing, because an
				// empty explanation reads as a file with nothing to say.
				fmt.Printf("\nexplain: %s is not in the program, so there is nothing to explain\n", *explainFile)
			} else {
				explanation, err := graph.Explain(ctx, subject, rules)
				if err != nil {
					return fmt.Errorf("explaining %s: %w", *explainFile, err)
				}
				printExplanation(os.Stdout, explanation)
			}
		}

		if reusedWalk {
			// Zero elapsed on purpose: the walk's cost belongs to the phase that performed it, and
			// recording it here too would bill one walk to two rows.
			report.record(phaseLint, outcomeReused, 0, "the fix phase's walk (nothing was rewritten)")
		} else {
			report.record(phaseLint, outcomeRan, lintDuration, "")
		}
	}

	// Phase 5: unused. A report, run only when asked for, and never a reason to fail a build.
	if !runUnused {
		report.record(phaseUnused, outcomeSkipped, 0, "not requested — this is a report, ask for it with --unused")
	} else {
		unusedStart := time.Now()

		// The honest statement about what the findings are worth. The reference analysis is keyed on
		// symbol identity, so unresolved symbols read as "nothing references this" and a broken build
		// produces a longer report rather than an error. Saying so beside the findings is the whole
		// difference between a report a reader can calibrate and one that quietly misleads.
		//
		// It is a warning rather than a refusal because the reachability half needs no types at all
		// and is fully trustworthy either way, so refusing to run would withhold good findings over a
		// caveat that only applies to the other half.
		if !runTypes {
			fmt.Printf(
				"\nunused: types did not run, so the reference half of this report is UNRELIABLE — " +
					"unresolved symbols look exactly like unreferenced ones. The unreachable half " +
					"below needs no types and is unaffected.\n",
			)
		}

		unusedResult, err := unused_exports.Run(ctx, graph, projectFiles, *unusedDeep)
		if err != nil {
			return fmt.Errorf("running the unused report: %w", err)
		}
		unusedDuration := time.Since(unusedStart)

		unused_exports.Write(os.Stdout, unusedResult, *unusedAll)

		// Deliberately NOT added to `findings`. The exit code is the gate's verdict, and this phase
		// is not part of the gate: an export held for an external consumer is unused and correct, so
		// failing a build over it would make the report something people route around rather than
		// read.
		report.record(phaseUnused, outcomeRan, unusedDuration, "")
	}

	// The phase line prints on every run, success included. A run that checked nothing must not be
	// able to print like a run that checked everything and found it clean, and a phase summary that
	// only appeared on failure would reintroduce exactly that ambiguity for the successful case.
	writeRunEnd(report, os.Stdout)

	// A crash fails the run as a finding does, though it is not one: see runCrash.
	if findings > 0 || activeSummary.Gaps.CrashedFiles > 0 || activeSummary.Gaps.RuleCrashes > 0 {
		finishRunCache(1)
	}
	return nil
}

// rebuildGraph reconstructs the type graph after files on disk have changed.
//
// The graph is a snapshot of the bytes as they were when it was built. Once the fix phase rewrites
// a file, every later phase reading that graph is reading source that no longer exists — the same
// stale-read corruption the edit engine refuses internally, one level up and with a wider blast
// radius, because a type diagnostic against deleted text points at a line nobody can find.
//
// It costs a full rebuild, which is why it only happens when something actually changed.
func rebuildGraph(
	configFileName string,
	directory string,
	singleThreaded bool,
	lintConfig *configuration.Config,
	yieldedFiles map[string]struct{},
) (*program.Graph, time.Duration, error) {
	start := time.Now()
	// The check's stats predate the rewrite. See distrustCheckStats.
	distrustCheckStats()
	rebuilt, err := program.Build(program.Options{
		ConfigFileName:   configFileName,
		CurrentDirectory: directory,
		SingleThreaded:   singleThreaded,
		ContentPack:      openContentPack(directory),
		Yielded:          yieldedFiles,
	})
	if err != nil {
		return nil, time.Since(start), err
	}
	rebuilt.LintConfig = lintConfig
	rebuilt.RuleOptions = registry.OptionsAt(optionsBase(lintConfig, directory))
	rebuilt.RegisteredRuleNames = registry.Names()
	rebuilt.Readiness = requestedReadiness()
	return rebuilt, time.Since(start), nil
}

// optionsBase is where a rule's path options are anchored: the lint config's own directory for a
// relative path, and the project root this run discovered for an absent one.
//
// Built at both places a graph receives its options, so the graph rebuilt after fixing anchors the
// same way the first one did. A rebuilt graph that decoded with no base would refuse a relative root
// the first graph accepted, and fail the run halfway through for a reason the config never changed.
// configureLint loads the lint config and hands it to the graph, which is what fixing and linting
// both need before either runs: which rules apply to which files, with their options. One function
// because the gate and the editor's save (stdin.go) must configure rules the same way, or a save would
// write what the gate does not.
//
// ahead, when not nil, is the same config already read beside the graph build (lintConfigAhead), and its
// load and settings check are taken from it rather than done again here.
func configureLint(graph *program.Graph, location projectLocation, ahead *lintConfigAhead) (*configuration.Config, error) {
	// A config that cannot be read is a hard failure and never a permissive default. Linting
	// everything with nothing configured produces output indistinguishable from a clean run, and
	// that exact confusion is what this tool exists to make impossible.
	var lintConfig *configuration.Config
	var err error
	if ahead != nil {
		<-ahead.done
		lintConfig, err = ahead.config, ahead.loadError
	} else {
		lintConfig, err = loadLintConfig(graph, location)
	}
	if err != nil {
		return nil, fmt.Errorf("loading the lint config: %w", err)
	}
	// Validated against the whole program rather than against a narrowed scope. The question
	// this check asks is whether a config override reaches any file at all, which is a property
	// of the tree and not of one run: scoping to a single file legitimately leaves `modules/**`
	// matching nothing, and failing there would make every scoped run report a broken config.
	//
	// Measured while building the scope: `--lint OneFile.ts` refused to run, naming three
	// overrides as vacuous, because the validator was handed the one file in scope.
	wholeProgramFiles := graph.ProjectFiles()
	projectFileNames := make([]string, 0, len(wholeProgramFiles))
	for _, projectFile := range wholeProgramFiles {
		projectFileNames = append(projectFileNames, projectFile.FileName().AsString())
	}
	if err := lintConfig.ValidateSelectors(projectFileNames); err != nil {
		return nil, fmt.Errorf("validating the lint config: %w", err)
	}
	base := optionsBase(lintConfig, location.Root)
	if ahead != nil {
		if ahead.settingsError != nil {
			return nil, fmt.Errorf("validating the lint config: %w", ahead.settingsError)
		}
		graph.RuleOptions = ahead.options
	} else {
		if err := registry.CheckSettings(base); err != nil {
			return nil, fmt.Errorf("validating the lint config: %w", err)
		}
		graph.RuleOptions = registry.OptionsAt(base)
	}
	graph.LintConfig = lintConfig
	graph.RegisteredRuleNames = registry.Names()
	graph.Readiness = requestedReadiness()
	return lintConfig, nil
}

func optionsBase(lintConfig *configuration.Config, projectRoot string) rule.OptionsBase {
	base := rule.OptionsBase{ProjectRoot: projectRoot}
	if lintConfig != nil {
		base.ConfigDirectory = lintConfig.Root
		base.Settings = lintConfig.Settings
	}
	return base
}

// collectTypeDiagnostics gathers the compiler's own findings for our files.
//
// The program is checked as a whole and then filtered down, rather than asked file by file. That
// ordering is load-bearing for speed and not at all for correctness: see Graph.AllDiagnostics. It
// does mean the compiler also checks the 6,500 third-party declarations, which is unavoidable — they
// have to be checked for our files to mean anything — and their findings are dropped here because
// nobody working in this repo can act on them.
//
// When the tsconfig is incremental, the check runs warm: files whose contents have not changed
// since the last run are not re-checked. Measured on the ahra tree, 1149ms cold against 380ms
// warm, with the finding set identical in both directions. A tsconfig without `incremental`
// falls back to the full pass, which is also what happens on the first run, after a toolchain
// version change, and any time the build info cannot be read.
//
// The build info is written even when this run reports findings, and that is deliberate: it
// records which files were checked, not whether they passed. A file that type-checked and
// produced a diagnostic was still checked. Discarding that on failure would make the slow path
// the one a person hits while iterating on an error, which is exactly the loop where the pause
// costs most.
//
// incremental is false under `--no-cache`, and then the build info is neither read nor written: the
// whole program is checked, as it is for a tsconfig that is not incremental.
//
// persist is false under `--no-fix`. The build info is still read, so the run is as warm as the last
// writing run left it, and only the write is withheld. A `--no-fix` run is therefore warm against a
// slightly older snapshot rather than cold, which costs it the files changed since then and nothing
// else; reading costs nothing the promise forbids, and refusing to read would make every CI run cold
// for no reason.
func collectTypeDiagnostics(ctx context.Context, graph *program.Graph, files []*ast.SourceFile, incremental bool, persist bool) []*ast.Diagnostic {
	return finishTypeCheck(ctx, startTypeCheck(ctx, graph, incremental), files, persist)
}

// typeCheck is TypeScript's own checking of a graph, under way or done.
type typeCheck struct {
	graph   *program.Graph
	session *program.IncrementalSession
	checked []*ast.Diagnostic
	done    chan struct{}

	// reuse is the types section of the run cache's table, nil when this run is not recorded. replayed is
	// how many files' semantic diagnostics the check took from it rather than checking.
	reuse    *program.TypeDiagnosticsReuse
	replayed int

	// elapsed is how long the checking itself took, and finished when it ended, both set as it finishes. The
	// phase reports these rather than its wait, which is nothing when the check ran alongside the walk.
	elapsed  time.Duration
	finished time.Time
}

// startTypeCheck begins checking every file, and returns at once.
//
// It runs alongside the fix phase's walk rather than after it. Both use the same checkers, each file under
// its own checker's lock, so on one checker the two take turns while different checkers go on in parallel:
// the run ends when the slowest checker has done its walk and its check, rather than after the slowest
// walk and then the whole check. Traced on ahra (#zqsdzbq), the check was a separate 0.1 to 0.2s after a
// walk whose slowest checker set its end.
//
// Only the per-file checking starts early. What reads every checker at once, the global and config
// diagnostics, waits for finishTypeCheck, when the walk has let go of them.
//
// Nothing ever cancels it. The compiler's whole-program check is not cancellable part way: a checker that
// sees its context end marks itself cancelled and the compiler hands it the next file anyway, where it
// panics on a goroutine nothing here can recover. A check whose graph is replaced is retired instead (see
// program.Graph.Retire), left to finish, and its result dropped.
func startTypeCheck(ctx context.Context, graph *program.Graph, incremental bool) *typeCheck {
	checkContext := ctx
	check := &typeCheck{graph: graph, done: make(chan struct{}), reuse: activeTypesReuse()}
	// One session across check-then-write. The build info has to be emitted from the same
	// incremental program that did the checking: that program's snapshot is what records which
	// files were checked, and emitting from a second one writes a build info that skips nothing
	// while looking correct. See program.IncrementalSession.
	go func() {
		// A test instrument: hold the check until the fix phase and any rebuild are over, so a test can put
		// the rebuild ahead of the check every time rather than by luck. See releaseEarlyTypeCheck.
		if os.Getenv("COHERE_TEST_HOLD_EARLY_CHECK") != "" {
			<-earlyTypeCheckReleased
		}
		start := time.Now()
		defer close(check.done)
		defer func() { check.elapsed, check.finished = time.Since(start), time.Now() }()
		// Begun here rather than before, since reading the build info and snapshotting the program is work
		// too, and it overlaps the walk as well as the check does.
		// Files whose shape fingerprint has not moved replay their semantic diagnostics, and only the rest are
		// checked. See program.TypeDiagnosticsReuse.
		if check.reuse != nil {
			if parts, reused := graph.CheckReusing(checkContext, check.reuse); reused {
				check.checked, check.replayed = parts.All(), check.reuse.Replayed()
				return
			}
		}
		if incremental {
			check.session = graph.NewIncrementalSession()
		}
		var parts program.TypeDiagnosticParts
		switch {
		case graph.FusedCheck != nil && check.session == nil:
			// The walk checked the files it visited, and this checks the rest.
			parts = graph.FusedCheck.DiagnosticParts(checkContext, graph)
		case check.session != nil:
			parts = check.session.DiagnosticParts(checkContext)
		default:
			parts = graph.AllDiagnosticParts(checkContext)
		}
		check.checked = parts.All()
		if check.reuse != nil {
			graph.RecordFull(checkContext, check.reuse, parts)
		}
	}()
	return check
}

// earlyTypeCheckReleased is closed once the fix phase and any rebuild are over, for COHERE_TEST_HOLD_EARLY_CHECK.
var (
	earlyTypeCheckReleased    = make(chan struct{})
	releaseEarlyTypeCheckOnce sync.Once
)

// releaseEarlyTypeCheck lets a held check start. Called when the fix phase is over, whether or not anything
// was held.
func releaseEarlyTypeCheck() {
	releaseEarlyTypeCheckOnce.Do(func() { close(earlyTypeCheckReleased) })
}

// finishTypeCheck waits for the check, writes the build info if asked, and returns the diagnostics that
// belong to files, plus the program's own.
func finishTypeCheck(ctx context.Context, check *typeCheck, files []*ast.SourceFile, persist bool) []*ast.Diagnostic {
	<-check.done

	// The read already happened when the session began; only the write is conditional.
	if check.session != nil && persist {
		if writeDiagnostics := check.session.Write(ctx); len(writeDiagnostics) > 0 {
			// A failed build-info write must not pass silently. The next run would be cold while
			// this one reported success, and the symptom is a saving that quietly never appears.
			for _, diagnostic := range writeDiagnostics {
				fmt.Fprintf(os.Stderr, "cohere: writing the incremental cache: %s\n",
					diagnostic.MessageKey())
			}
		}
	}

	ours := make(map[*ast.SourceFile]struct{}, len(files))
	for _, sourceFile := range files {
		ours[sourceFile] = struct{}{}
	}
	diagnostics := check.graph.ConfigDiagnostics(ctx)
	if check.reuse != nil {
		check.reuse.FinishGlobals(len(check.graph.Program.GetGlobalDiagnostics(ctx)) == 0)
	}
	for _, diagnostic := range check.checked {
		// A diagnostic with no file is about the program rather than about any one file, so it is ours
		// by default: dropping it would hide exactly the configuration errors that matter most.
		if sourceFile := diagnostic.File(); sourceFile != nil {
			if _, isOurs := ours[sourceFile]; !isOurs {
				continue
			}
		}
		diagnostics = append(diagnostics, diagnostic)
	}
	return diagnostics
}

// printCompilerDiagnostic prints one of TypeScript's own findings, in the shape tsc uses so an
// editor's problem matcher recognizes it.
func printCompilerDiagnostic(diagnostic *ast.Diagnostic) {
	sourceFile := diagnostic.File()
	if sourceFile == nil {
		printFinding(os.Stdout, compilerFinding(diagnostic), fmt.Sprintf("error TS%d: %s\n", diagnostic.Code(), diagnosticMessage(diagnostic)))
		return
	}

	line, character := scanner.GetECMALineAndByteOffsetOfPosition(sourceFile, diagnostic.Loc().Pos())
	printFinding(os.Stdout, compilerFinding(diagnostic), fmt.Sprintf(
		"%s:%d:%d - error TS%d: %s\n",
		sourceFile.FileName(), line+1, character+1, diagnostic.Code(), diagnosticMessage(diagnostic),
	))
}

// diagnosticMessage renders a compiler diagnostic's text.
//
// `MessageText()` is the wrong accessor and returns the empty string for every diagnostic the
// checker produces. A compiler diagnostic carries a message template and its arguments separately,
// and `messageText` is only populated for external diagnostics that arrive pre-localized. So a real
// type error printed as `error TS2322: ` with nothing after the colon, which is worse than not
// printing it: the reader is told a file is broken and not told how.
//
// `Localize` is what the compiler's own diagnostic writer uses. It renders the template with its
// arguments, and falls back to `messageText` for the external case, so it is correct for both rather
// than for the one that happened to be tested.
//
// The zero Locale selects the built-in English text, which is what this tool prints everywhere else.
func diagnosticMessage(diagnostic *ast.Diagnostic) string {
	return diagnostic.Localize(locale.Locale{})
}

// singleLineDescription collapses a message's internal newlines so one finding prints as one line.
//
// Named rather than inlined so a test can hold it. A test asserting `strings.ReplaceAll` on its own
// input passes whatever the printer does, which is a test of the standard library rather than of
// this program.
func singleLineDescription(description string) string {
	return strings.ReplaceAll(description, "\n", " ")
}

// printRuleDiagnostic prints one rule finding in the same shape, with the rule name where the error
// code goes — a reader should not have to learn two formats. The settings give it its severity.
func printRuleDiagnostic(out io.Writer, diagnostic rule.Diagnostic, lintConfig *configuration.Config) {
	sourceFile := diagnostic.SourceFile
	if sourceFile == nil {
		printFinding(out, ruleFinding(diagnostic, lintConfig), fmt.Sprintf("error %s: %s\n", diagnostic.RuleName, diagnostic.Message.Description))
		return
	}

	line, character := scanner.GetECMALineAndByteOffsetOfPosition(sourceFile, diagnostic.Range.Pos())

	// One finding is one line, whatever the message does internally.
	//
	// Two rules write a second paragraph into their description with an embedded newline, which put
	// the rule tag on a line carrying no position. Any reader keyed on `position ... [rule]` then
	// drops those findings silently, and it drops the LONGEST messages, which are the ones carrying
	// advice. Four of 5,312 findings on the ahra tree wrapped that way, and an extractor that missed
	// them sent me four hours into a rule that was working correctly.
	//
	// Joining rather than forbidding the newline, because the rules are not wrong to want a second
	// paragraph: the message is the rule's to write and the line discipline is the printer's to keep.
	description := singleLineDescription(diagnostic.Message.Description)

	printFinding(out, ruleFinding(diagnostic, lintConfig), fmt.Sprintf(
		"%s:%d:%d - %s [%s/%s]\n",
		sourceFile.FileName(), line+1, character+1,
		description, diagnostic.RuleName, diagnostic.Message.Id,
	))
}

// changedConfiguration names the configuration file in a changed-file scope, or empty when none is.
//
// Three kinds, and all three reach every file at once: the rule config, the tsconfig that built the
// program, and every tsconfig that one extends. The extends chain matters here rather than being a
// completeness gesture. On this tree the root tsconfig holds one line, `extends`, and every option
// and every include pattern lives in the base it points at, so a guard checking only the root would
// miss every real change to how the program is built.
//
// Dependencies are deliberately not a fourth kind, and the reason is worth stating because
// `package.json` looks like the obvious next entry. It is a declaration; the checker reads
// `node_modules`, and the two move independently: a version edited with no install changes nothing
// the program can see, and an install changes every consumer of a package without touching a
// tracked file. `node_modules` is gitignored, so git cannot report the change that actually
// matters. A guard on `package.json` would fire on the edits that mean nothing and stay silent on
// the ones that mean everything, which is worse than the gap it appears to close.
//
// So a dependency change is a real limit of `--changed` rather than a defect in it. The honest
// answer after an install is a whole-tree run, and the coverage line saying how many files were
// checked is what leaves that visible.
//
// Its own function so a fixture can hold the comparison rather than reconstruct it. The
// reconstruction is what makes this untestable in place: asserting that `resolveLintConfigPath`
// plus `Clean` matches an absolute scope entry is a true statement about two helpers and says
// nothing about whether the call site passes them the right directory.
//
// It passed the wrong one first. `--directory` defaults to empty, so resolving against the flag
// left a bare relative name that never matched an absolute scope entry, and the guard silently
// never fired. Resolving against the graph's directory is what makes it fire.
func changedConfiguration(scope formatScope, graph *program.Graph, lintConfigFileName string) string {
	currentDirectory := ""
	if graph.Config != nil {
		currentDirectory = graph.Anchor.Root()
	}

	// The lint config resolves against the graph's directory when there is one, and against the
	// tsconfig's own directory otherwise. Both are the tree root in practice, and falling back to the
	// empty string would leave a relative path that matches nothing, which is how the first version
	// of this guard was silently inert.
	if currentDirectory == "" && graph.ConfigFileName != "" {
		currentDirectory = filepath.Dir(graph.ConfigFileName)
	}

	candidates := append(configuration.SourcesOnDisk(lintConfigSources(resolveLintConfigPath(lintConfigFileName, currentDirectory))), graph.ConfigFileName)
	if graph.Config != nil {
		for _, extended := range graph.Config.ExtendedSourceFiles() {
			candidates = append(candidates, extended.AsString())
		}
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if scope.includes(filepath.Clean(candidate)) {
			return candidate
		}
	}
	return ""
}

// writeResolvedConfig prints every registered rule's resolved setting for one file, as JSON in the shape
// ESLint's `--print-config` uses: `{"rules": {"name": [severity, ...options]}}`, severity 0, 1 or 2.
//
// Resolved per registered rule by the resolver's own question, not read from the config's keys. A key
// can configure a rule spelled differently (`@typescript-eslint/no-invalid-this` configures core
// `no-invalid-this` when no key names it exactly), so the keys are the wrong answer to "what runs",
// which is the answer a parity check against ESLint needs (#mnmx9s4). A rule nobody configured is
// left out, as ESLint leaves out a rule no config names.
func writeResolvedConfig(out io.Writer, resolved configuration.Resolved) error {
	rules := map[string][]json.RawMessage{}
	for _, registered := range registry.All() {
		status, setting := resolved.StatusOf(registered.Name)
		var severity int
		switch status {
		case configuration.StatusEnabled:
			severity = int(setting.Severity)
		case configuration.StatusScopedOff:
			severity = 0
		default:
			continue
		}
		entry := []json.RawMessage{json.RawMessage(strconv.Itoa(severity))}
		entry = append(entry, setting.Options...)
		rules[registered.Name] = entry
	}
	encoded, err := json.MarshalIndent(map[string]any{"rules": rules}, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(encoded))
	return err
}

// registeredRuleNames is every rule this binary runs, by name, which the config loader needs to tell a
// key spelled differently from the one it inherits apart from a twin rule's key.
func registeredRuleNames() []string {
	all := registry.All()
	names := make([]string, 0, len(all))
	for _, registered := range all {
		names = append(names, registered.Name)
	}
	return names
}

// lintConfigSources is the lint config and every source it extends, or the config alone when the chain
// cannot be read. An embedded set is named as `cohere:<name>`, which no file is; a caller that stats
// what it is handed takes configuration.SourcesOnDisk of it.
//
// The fallback is not a way past a broken chain: configureLint loads the same chain and fails the run,
// so a run with an unreadable base is never recorded or replayed. It only keeps the callers here,
// which name inputs, from naming none.
func lintConfigSources(path string) []string {
	sources, err := configuration.SourcesOf(path)
	if err != nil || len(sources) == 0 {
		return []string{path}
	}
	return sources
}

// resolveLintConfigPath finds the lint config relative to the directory paths resolve against.
//
// A relative config path resolved against the process's own working directory rather than the
// project's would quietly find nothing, and "nothing" here means every rule runs everywhere.
func resolveLintConfigPath(configPath string, directory string) string {
	if filepath.IsAbs(configPath) || directory == "" {
		return configPath
	}
	return filepath.Join(directory, configPath)
}

// writeOrphanedConfigKeys says which config entries name a rule that does not exist. Printed in full
// by default, because a decision that silently stopped applying needs action.
//
// This is the inverse of the not-configured count, and it is the failure the rename of every
// rule to its upstream spelling introduced. The resolver matches a config key exactly, or trims a
// prefix off it on a `/` boundary, so a key that is LONGER than the rule name resolves and a key
// that is SHORTER never can. `typescript/no-base-to-string` cannot reach a rule registered as
// `@typescript-eslint/no-base-to-string`: there is no prefix to trim, only one to add.
//
// Left silent, this reads as the opposite of what happened. The rule reports "nobody has said
// whether it should run" while someone did say, in the config, in a key sitting three lines above
// one that works. A deliberate `off` that stops applying is worse than one that was never written,
// because the config still shows the decision and the run no longer honors it.
//
// Keys naming a rule cohere has not ported yet are the expected case and not an error: a decision
// recorded ahead of the rule is how this migration is supposed to work. They are counted and named
// separately so the two are never confused.
func writeOrphanedConfigKeys(out io.Writer, rules []rule.Rule, config *configuration.Config) {
	if config == nil {
		return
	}

	registered := make(map[string]bool, len(rules))
	for _, registeredRule := range rules {
		registered[registeredRule.Name] = true
	}

	// A key resolves if some registered rule matches it exactly, or ends with it on a `/` boundary.
	// This mirrors settingFor rather than reimplementing it loosely, because a check that disagreed
	// with the resolver would report keys that work and miss keys that do not.
	resolves := func(key string) bool {
		for name := range registered {
			if configuration.KeyReachesRule(key, name) {
				return true
			}
		}
		return false
	}

	orphaned := make([]string, 0)
	for key := range config.Rules {
		if !resolves(key) {
			orphaned = append(orphaned, key)
		}
	}
	sort.Strings(orphaned)

	for _, key := range orphaned {
		setting := config.Rules[key]
		fmt.Fprintf(out,
			"  config: key %q matches no registered rule, so its %s never applies — either the rule is not ported yet, or the key is spelled for an older name\n",
			key, setting.Severity,
		)
	}
}

// round trims a duration to milliseconds, which is the resolution any of these numbers means
// anything at.
func round(duration time.Duration) time.Duration {
	return duration.Round(time.Millisecond)
}

// implementsConfiguredRule reports whether a config entry names a rule this binary contains.
func implementsConfiguredRule(configured string, implemented map[string]bool) bool {
	if implemented[configured] {
		return true
	}
	if slash := strings.LastIndexByte(configured, '/'); slash >= 0 {
		return implemented[configured[slash+1:]]
	}
	return false
}

// filterToScope keeps the project files a scope names, in their original order.
//
// Order is preserved rather than rebuilt from the scope, because the walk reports findings in the
// order it receives files and a scoped run should print what an unscoped run printed for those same
// files, in the same sequence. Sorting here would make the two disagree for no reason a reader
// could see.
func filterToScope(projectFiles []*ast.SourceFile, scope formatScope) []*ast.SourceFile {
	kept := make([]*ast.SourceFile, 0, len(scope.FileNames))
	for _, sourceFile := range projectFiles {
		if scope.includes(filepath.Clean(sourceFile.FileName().AsString())) {
			kept = append(kept, sourceFile)
		}
	}
	return kept
}

// rescopeAfterRebuild carries a scoped file set onto a rebuilt graph.
//
// Matched by name, because a rebuild parses every file again and the old source file pointers
// belong to a graph nothing reads any more. The set is the one already decided, closure included,
// rather than a closure recomputed on the new graph: the scope is what the run told the reader it
// would check, and the graph line has already printed its size.
//
// A scoped file the rebuilt program no longer holds is returned by name rather than dropped, so a
// file that left the scope says so instead of shrinking the count without a word.
func rescopeAfterRebuild(rebuilt []*ast.SourceFile, scoped []*ast.SourceFile, everything bool) ([]*ast.SourceFile, []string) {
	if everything {
		return rebuilt, nil
	}

	wanted := make(map[string]bool, len(scoped))
	for _, sourceFile := range scoped {
		wanted[sourceFile.FileName().AsString()] = false
	}

	kept := make([]*ast.SourceFile, 0, len(scoped))
	for _, sourceFile := range rebuilt {
		if _, inScope := wanted[sourceFile.FileName().AsString()]; inScope {
			wanted[sourceFile.FileName().AsString()] = true
			kept = append(kept, sourceFile)
		}
	}

	lost := []string{}
	for fileName, found := range wanted {
		if !found {
			lost = append(lost, fileName)
		}
	}
	sort.Strings(lost)
	return kept, lost
}

// narrowToClosure turns a scope into the files a run should visit, and says when it declined.
//
// Shared by the named-path and changed-file paths rather than written at each, because the judgment
// is the same and it is the judgment that is easy to get wrong. The files a caller pointed at are
// not the answer on their own: a change to an exported type breaks its consumers, and those
// findings vanish rather than appear. Demonstrated by giving one function an `any` return, which
// produced seven new findings across the tree, four of them in a consumer two directories from the
// edit, and a run scoped to the edited file alone saw three.
//
// Above the closure limit the whole tree is checked and a note says why. A truncated closure would
// be fast and silently miss findings, which is worse than being slow, and it is the one failure a
// fast path must not introduce.
func narrowToClosure(
	graph *program.Graph,
	scope formatScope,
	projectFiles []*ast.SourceFile,
) (formatScope, []*ast.SourceFile) {
	if scope.Everything {
		return scope, projectFiles
	}

	seeds := filterToScope(projectFiles, scope)
	closure, within := graph.DependentClosure(seeds)
	if !within {
		fmt.Fprintf(accountOutput(os.Stderr),
			"note: %s reaches more than %d files through imports, so the whole tree is checked\n",
			scope.RequestDescription, program.DependentClosureLimit)
		return formatScope{Everything: true}, projectFiles
	}

	scope.DependentCount = len(closure) - len(seeds)
	return scope, closure
}

// reportForeignCheckers says, under the COHERE_TEST_FOREIGN_CHECKERS instrument, how many files a walk put on
// a checker other than their own, so a test comparing the two runs can tell the instrument took effect.
func reportForeignCheckers(graph *program.Graph, walked program.Result, files int) {
	if graph.WalkOnForeignCheckers {
		fmt.Fprintf(os.Stderr, "foreign checkers: %d of %d files walked on a checker other than their own\n",
			walked.FilesOnForeignCheckers, files)
	}
}

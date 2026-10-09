// The run-scoped design system: built once per program, shared read-only across the worker pool.
//
// This is the seam between `internal/tailwind`, which knows how to build a design system and
// nothing about when, and the rules, which need one on every file and must not each build their
// own. No rule reads it yet — swapping the rules over is #4q5dsn3 and #3r6cxrb. What lands here is
// the seam and the proof that it is built once.
//
// # The pattern, and the failure history it carries
//
// Copied deliberately from `internal/rules/structure/boundary_no_project_theme_value.go`, whose own
// comment records why each part is shaped the way it is:
//
//   - **Keyed on the program pointer, not on nothing.** That rule originally memoized in a
//     module-level variable with no key, which is correct for a process that lints once and exits
//     and wrong for anything that lints twice: the second tree reads the first's data. Pointer
//     identity gives a miss on a new program for free. `TestDesignSystemCacheKeyIsTheProgram`
//     mutates the key away and asserts the bug reappears, so the key is load-bearing rather than
//     decorative.
//   - **A package variable rather than `rule.FileCache`.** That cache is per-file by construction,
//     its own comment says sharing an entry across files would be a bug, and it carries no mutex.
//     It cannot live at registration either, since the program is built after rules register.
//   - **The mutex sits at the only access point and the key is immutable for the run**, so this is
//     safe under the parallel walk by construction rather than by discipline.
//
// # Why the build-once property is a test
//
// That same rule measured 2,329ms of setup against 0.54ms of listening across 1,862 files, 64.5% of
// all rule time for a rule that reported nothing, because the scan was redone per file. Nothing
// failed; it was only slow, and the comment still said "once per run".
//
// The design system build is the same shape. Measured on both corpus repositories: 0.68ms per build
// on ~/Projects/ahra, over a 9-file `@import` graph resolving 744 theme entries and 37 `@utility`
// blocks, and 0.64ms on www-connected-app for 748 entries. Per-file across 1,862 files that is 1.27
// seconds, and across the 3,416-file tree the rule catalog is measured on, about 2.3 seconds —
// against a lint phase that currently runs in 0.19s. The tool would get twelve times slower and
// every test would stay green.
//
// So `TestDesignSystemIsBuiltOncePerProgram` walks a real program with a rule that asks for the
// design system on every file, and asserts the process-wide build counter moved by exactly one.
// A counter is the only form of that claim that cannot decay.
//
// # Failing safe
//
// A design system that will not build is no opinion, never a clean tree. The load result carries
// its error and `DesignSystemForProgram` hands it back on every call, so a rule that cannot get one
// must decline and say so. Returning a zero-valued system instead would make every rule report
// nothing on a repository whose CSS moved, which is confident green over unchecked work and the
// exact failure cohere exists to remove.
package tailwind

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rule"
	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
)

// DesignSystemResult is one program's design system, or the reason there is none.
//
// Both are cached, and caching the failure is the point rather than an oversight. A repository with
// no Tailwind entry point would otherwise pay a full filesystem search on every one of its files to
// rediscover that it has none, which is the per-file cost this whole file exists to refuse, in the
// case where there is not even a design system to show for it.
type DesignSystemResult struct {
	// System is the loaded design system, nil when Err is set.
	System *tailwindengine.LoadedDesignSystem
	// Table is this design system's descriptor table, nil when Err is set.
	//
	// Built here rather than by each rule, and that is the same claim the whole file makes about the
	// design system: a rule computing a sort key needs a table on every class in the repository, and
	// a table built per rule or per file is the per-file rebuild this file exists to refuse, wearing
	// a different name. It costs a few map copies over the framework's roots plus one PropertySort
	// per static `@utility` block, on a path the build counter already guards.
	Table *tailwindengine.Table
	// Err is why no design system could be built, nil when System is set.
	//
	// A rule holding this must decline rather than report zero findings. See the failing-safe note
	// in the file comment.
	Err error
	// EntryPoint is the stylesheet the load was attempted from, empty when none was found.
	EntryPoint string
	// Reads is every path the load asked the file system about, present or absent: the entry point
	// candidates, the package root's probes, and every stylesheet in the `@import` graph. It is the
	// design system's whole read set, observed through rule.RecordingFS rather than listed, so a cache
	// keyed on it re-runs the Tailwind rules when any of those files changes, appears, or disappears
	// (#pyhm2t2). Set on every result, a failed load's included: a missing entry point is an input too.
	Reads []rule.FileRead
}

// ErrNoTailwindEntryPoint is returned when a program's project holds no Tailwind stylesheet.
//
// A distinct error rather than a generic one because the two states a caller must tell apart are
// "this project does not use Tailwind" and "this project uses Tailwind and its CSS would not
// parse". The first is silence with nothing wrong; the second is a rule that must decline loudly.
var ErrNoTailwindEntryPoint = errors.New("no Tailwind entry point found in this project")

// designSystemCache holds the design system for one program, so the build happens once per run.
//
// Keyed on the program pointer rather than on nothing, which is what makes this safe to reuse
// across trees: a second program is a miss for free. A package variable rather than
// `rule.FileCache`, which is per-file by construction and carries no mutex. The mutex is at the
// only access point and the key is immutable for the run, so the parallel walk is safe by
// construction. Every one of those three choices has the failure that motivated it recorded in the
// file comment above.
var designSystemCache struct {
	sync.Mutex
	program rule.ProgramIdentity
	// results holds one design system per location the rules' options name in this program. Nearly
	// every project names none, so this holds the one default entry.
	results map[DesignSystemLocation]DesignSystemResult
}

// DesignSystemLocation is where a rule's options say the design system is: upstream's `cwd`,
// `entryPoint` and `tailwindConfig`, with the directory a relative path anchors to. The zero value is
// the default, the project's own stylesheet found by FindEntryPoint.
type DesignSystemLocation struct {
	// Anchor is upstream's `ctx.cwd`, the directory ESLint runs from, which is the config file's
	// directory here. Empty means the project root.
	Anchor string
	// Cwd is upstream's `cwd`, as written: where Tailwind and the stylesheet are looked for from,
	// relative to Anchor when relative.
	Cwd string
	// ConfigPath is upstream's `entryPoint ?? tailwindConfig`, as written.
	ConfigPath string
}

// isDefault reports whether no option named a location, so the project's own stylesheet is used.
func (location DesignSystemLocation) isDefault() bool {
	return location.Cwd == "" && location.ConfigPath == ""
}

// DesignSystemForProgram returns this run's design system, building it at most once.
//
// Every rule that calls this must declare `ReadsCompilerOptions | ReadsDesignSystem`. The program reaches every file in
// the run and the stylesheet graph reaches files the program does not contain at all, so a findings
// cache keyed on the linted file alone is stale whenever `theme.css` changes and the `.tsx` file
// does not: zero findings, forever, indistinguishable from a clean tree.
//
// It takes the program rather than the rule's Context so that the declaration is enforced rather
// than asked for. The guard in internal/guard is textual: a rule file that names
// `ctx.Program` must declare. When this took the Context, a rule could call it without ever writing
// `ctx.Program`, and the five rules that did passed the guard only because each also checked
// `ctx.Program == nil` somewhere else in the file (#ym4v8bc). Now the call itself is what the guard
// reads, and the guard refuses a helper outside a rule that reads `ctx.Program` on its callers'
// behalf, which is the shape this had.
//
// A nil program is a hard miss rather than a shared entry. The harnesses that build a Context by
// hand leave Program nil, and letting them share one cache slot would mean two unrelated fixtures
// reading each other's design system, which is the same bug the pointer key exists to prevent.
func DesignSystemForProgram(program rule.Program) DesignSystemResult {
	return DesignSystemForProgramAt(program, DesignSystemLocation{})
}

// DesignSystemForProgramAt is DesignSystemForProgram for the location a rule's options name, built at
// most once per program and location.
func DesignSystemForProgramAt(program rule.Program, location DesignSystemLocation) DesignSystemResult {
	if program == nil {
		return DesignSystemResult{Err: fmt.Errorf("no program: a design system cannot be located without one")}
	}

	designSystemCache.Lock()
	defer designSystemCache.Unlock()

	if designSystemCache.program != program.Identity() || designSystemCache.results == nil {
		designSystemCache.program = program.Identity()
		designSystemCache.results = map[DesignSystemLocation]DesignSystemResult{}
	}
	if result, isBuilt := designSystemCache.results[location]; isBuilt {
		return result
	}

	result := loadDesignSystemForProgram(program, location)
	designSystemCache.results[location] = result
	return result
}

// loadDesignSystemForProgram finds the project's stylesheet and builds a design system from it.
//
// Call `DesignSystemForProgram` rather than this: an uncached call re-walks the `@import` graph, and
// the rules that will read it run on every file in the tree.
func loadDesignSystemForProgram(program rule.Program, location DesignSystemLocation) DesignSystemResult {
	fileSystem := program.DesignSystemFS()
	result := loadDesignSystemThrough(program, fileSystem, location)
	result.Reads = fileSystem.Reads()
	return result
}

// loadDesignSystemThrough is the load itself, asking fileSystem every question about the disk.
func loadDesignSystemThrough(program rule.Program, fileSystem *rule.RecordingFS, location DesignSystemLocation) DesignSystemResult {
	projectRoot := projectRootOf(program)
	if projectRoot == "" {
		return DesignSystemResult{Err: fmt.Errorf("could not determine the project root from the program")}
	}
	if !location.isDefault() {
		return loadConfiguredDesignSystem(projectRoot, fileSystem, location)
	}

	// Every question about the disk goes through the program's own filesystem rather than to `os`,
	// because that filesystem is where the run cache's input recorder sits. A stylesheet read behind
	// its back is an input the cache never signs, so editing an ignored or untracked stylesheet the
	// theme imports replayed the old verdict (#ym4v8bc). Asked through it, each candidate entry point
	// probed and missed is recorded absent, so creating one invalidates, and each probe of the
	// package walk is recorded the same way. It is reached through a rule.RecordingFS, which also keeps
	// the read set the findings cache keys the Tailwind rules on.
	fileExists := func(path string) bool { return fileSystem.FileExists(tspath.RootedFilePath(path)) }

	entryPoint := FindEntryPoint(projectRoot, fileExists)
	if entryPoint == "" {
		return DesignSystemResult{Err: fmt.Errorf("%w: looked under %s", ErrNoTailwindEntryPoint, projectRoot)}
	}

	packageRoot := findTailwindPackageRoot(filepath.Dir(entryPoint), fileExists)
	if packageRoot == "" {
		return DesignSystemResult{
			EntryPoint: entryPoint,
			Err: fmt.Errorf(
				"%s: no installed tailwindcss package found, so `@import \"tailwindcss\"` cannot resolve",
				entryPoint,
			),
		}
	}

	system, err := tailwindengine.LoadDesignSystem(tailwindengine.LoadOptions{
		EntryPoint:          entryPoint,
		TailwindPackageRoot: packageRoot,
	})
	if err != nil {
		return DesignSystemResult{EntryPoint: entryPoint, Err: err}
	}
	// The stylesheets themselves are read by the engine, which does not take a filesystem, so each
	// one in the `@import` graph is stated through the program's afterwards. That records it as a
	// present input with its signature, which is all the cache needs: it re-signs the file on the
	// next run and misses when the file moved.
	for _, stylesheet := range system.Stylesheets {
		fileSystem.Stat(tspath.RootedPath(stylesheet))
	}
	// The descriptor table is built here, on the counted path, rather than lazily on first use. A
	// second build path would not be visible to TestDesignSystemIsBuiltOncePerProgram, and an
	// invisible build path is exactly how the per-file rebuild came back the last time.
	return DesignSystemResult{System: system, Table: tailwindengine.NewTable(system), EntryPoint: entryPoint}
}

// projectRootOf is the directory the program's tsconfig sits in.
//
// `ConfigFilePath` rather than `GetCurrentDirectory`, because the working directory is where cohere
// was invoked from and the project root is where its config lives, and the two differ whenever
// anyone runs the linter from a parent directory. Falls back to the current directory when the
// program was built without a config file, which is what the in-memory test harnesses do.
func projectRootOf(program rule.Program) string {
	if options := program.Options(); options != nil && options.ConfigFilePath != "" {
		return filepath.Dir(options.ConfigFilePath.AsString())
	}
	return program.GetCurrentDirectory()
}

// tailwindEntryPointCandidates are the paths a project's root stylesheet is looked for at,
// in order.
//
// A fixed list rather than a search, and that is the decision rather than a shortcut. A glob for
// `**/*.css` on a repository this size walks `node_modules` and every build output, which costs
// more than the design system it is trying to find and can pick a stylesheet nobody imports. Both
// repositories in this port's corpus put theirs at the first path; the rest are the conventional
// Tailwind 4 locations.
//
// A project that puts its stylesheet somewhere else gets ErrNoTailwindEntryPoint, which is a rule
// declining loudly rather than a rule reporting a clean tree. That is the right failure: adding a
// path here is a one-line change, and a wrong-stylesheet guess is a whole run of wrong answers.
var tailwindEntryPointCandidates = []string{
	filepath.Join("app", "_theme", "styles", "theme.css"),
	filepath.Join("app", "globals.css"),
	filepath.Join("src", "app", "globals.css"),
	filepath.Join("styles", "globals.css"),
	filepath.Join("src", "styles", "globals.css"),
	filepath.Join("app", "theme.css"),
}

// FindEntryPoint returns the project's root stylesheet, or empty when there is none.
//
// fileExists answers for files only, never directories. The design system asks the program's
// filesystem, so the run cache records each probe; see loadDesignSystemForProgram. Exported for zero
// config's detection of cohere:tailwind (internal/lint/housesets), which must find the same stylesheet
// the rules then read, so there is one search rather than two that could disagree.
func FindEntryPoint(projectRoot string, fileExists func(path string) bool) string {
	for _, candidate := range tailwindEntryPointCandidates {
		path := filepath.Join(projectRoot, candidate)
		if fileExists(path) {
			return path
		}
	}
	return ""
}

// diskFileExists is fileExists for callers with no program to ask, which is the tests.
func diskFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// findTailwindPackageRoot walks up from a directory looking for an installed `tailwindcss`.
//
// Upward from the stylesheet rather than from the project root, because that is how Node resolves
// and because a monorepo hoists: this repository's stylesheet is several directories below the
// `node_modules` that holds its Tailwind. The walk stops at the filesystem root.
func findTailwindPackageRoot(start string, fileExists func(path string) bool) string {
	directory := start
	for {
		candidate := filepath.Join(directory, "node_modules", "tailwindcss")
		if fileExists(filepath.Join(candidate, "index.css")) {
			return candidate
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return ""
		}
		directory = parent
	}
}

// DesignSystemDeclineMessage is what a rule says when it has no design system to work from.
//
// A shared message rather than one per rule, because every rule that reads the design system
// declines for exactly the same reasons and a reader hitting it in two rules should not have to
// decide whether they mean the same thing. It names the entry point when one was found, since "the
// CSS would not parse" and "there is no CSS" are different problems and the fix differs.
func DesignSystemDeclineMessage(ruleName string, result DesignSystemResult) string {
	var where string
	if result.EntryPoint != "" {
		where = " from " + result.EntryPoint
	}
	return fmt.Sprintf(
		"%s could not build this project's Tailwind design system%s, so it is reporting nothing rather "+
			"than reporting a clean tree: %v",
		ruleName, where, result.Err,
	)
}

// resetDesignSystemCacheForTest clears the cache so a test can measure a cold build.
//
// Test-only and named so, because a caller in the shipped path that reset this would reintroduce
// the per-file rebuild this file exists to prevent, and it would do it silently. Not in a _test.go
// file only because the tests that need it live in this package and the one that asserts the cache
// key needs to reach the key itself.
func resetDesignSystemCacheForTest() {
	designSystemCache.Lock()
	defer designSystemCache.Unlock()
	designSystemCache.program = rule.ProgramIdentity{}
	designSystemCache.results = nil
}

// stylesheetsUnder reports the design system's stylesheet graph as project-relative paths.
//
// For reporting and for tests. Absolute paths in a test failure message name a machine rather than
// a repository, and a failure a reader cannot place is one they cannot act on.
func stylesheetsUnder(root string, system *tailwindengine.LoadedDesignSystem) []string {
	if system == nil {
		return nil
	}
	relative := make([]string, 0, len(system.Stylesheets))
	for _, path := range system.Stylesheets {
		if trimmed, err := filepath.Rel(root, path); err == nil && !strings.HasPrefix(trimmed, "..") {
			relative = append(relative, trimmed)
			continue
		}
		relative = append(relative, path)
	}
	return relative
}

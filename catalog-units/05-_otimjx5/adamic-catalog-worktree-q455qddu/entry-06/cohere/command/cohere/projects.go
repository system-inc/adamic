package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/system-inc/cohere/internal/types/program"
)

// projectEngineVariable names the engine a child run checks its --directory with. A directory holding both
// a tsconfig.json and a Package.swift is two projects, and --directory alone reads it as TypeScript's, so
// the run that checks its Swift says which engine it is.
const projectEngineVariable = "COHERE_PROJECT_ENGINE"

// projectLabelVariable names the project a child run checks, by its path under the root, so the run's own
// footer can say which project it is about (#ytqqv8v).
const projectLabelVariable = "COHERE_PROJECT_LABEL"

// discoveryApplies reports whether a run discovers its projects. Anything that names one project, a
// directory, a tsconfig or a path, narrows the run to it as it always did, and so do the listings, the
// explanations and the editor's save, which answer about one project.
//
// A lint config does not: it says which rules apply, not which project, so every project discovered runs
// under it. It used to narrow, and the quiet hundred's harness, which names one, checked 2 of TanStack
// Query's thousand files and refused prisma outright (#wvgxtey).
func discoveryApplies(given map[string]bool, arguments []string) bool {
	if len(arguments) > 0 {
		return false
	}
	for _, name := range []string{
		"directory", "tsconfig", "rules", "rules-enabled", "print-config",
		"version", "cache-dump", "explain", "stdin-filepath", "profile",
	} {
		if given[name] {
			return false
		}
	}
	return true
}

// checkDiscoveredProjects discovers the projects under the root and, when there is more than the one the
// walk up located, checks them all and reports checked with the merged exit code. With only that one it
// prints what discovery passed over and leaves the run to check it in this process.
//
// lintConfigFileName is the lint config the command line named, absolute, or empty: every project's run
// gets it, since a path typed here would resolve against each project's directory there.
func checkDiscoveredProjects(location projectLocation, locateError error, workingDirectory string, lintConfigFileName string) (int, bool, error) {
	root := discoveryRoot(location, locateError, workingDirectory)
	patterns, err := rootIgnorePatterns(root)
	if err != nil {
		return 0, false, err
	}
	found, err := discoverProjects(root, patterns)
	if err != nil {
		return 0, false, err
	}
	if found.Ownership, err = resolveOwnership(&found); err != nil {
		return 0, false, err
	}
	if len(found.Projects) == 0 {
		// Nothing to check is a failure, never a green over nothing.
		solutions := ""
		if len(found.Ownership.Solutions) > 0 {
			solutions = fmt.Sprintf(" (%d solution tsconfigs found, %s, whose references reach no project here)",
				len(found.Ownership.Solutions), strings.Join(found.Ownership.Solutions, ", "))
		}
		return 0, false, fmt.Errorf("no %s or %s in %s, any directory above it, or any directory below it that the repository keeps, "+
			"so there is no project here to check%s (not searched: %d directories .gitignore ignores, %d dependency, build, cache "+
			"or fixture directories, %d nested repositories): run from inside one, or name it with -tsconfig or -directory",
			projectMarker, swiftProjectMarker, root, solutions, found.Ignored, found.NeverDescended,
			len(found.NestedRepositories)+found.Submodules)
	}
	if found.isTheLocatedProject(location, locateError) {
		// The project runs here, and its build lists the directories discovery just did.
		discoveredListings = found.Listings
		for _, note := range found.notes() {
			fmt.Fprintf(accountOutput(os.Stderr), "cohere: %s\n", note)
		}
		return 0, false, nil
	}
	return runProjects(found, childArguments(os.Args[1:], lintConfigFileName), os.Stdout), true, nil
}

// discoveredListings are the directory listings discovery read, when the one project it found runs in this
// process, for its build to serve the tsconfig's include enumeration from. Nil otherwise. See #bjv0tg4.
var discoveredListings *program.DirectoryListings

// childArguments are the arguments every project's run gets: this run's, with a named lint config made
// absolute. Discovery runs only when no path was named, so every argument is a flag, and a flag's value is
// either joined to it with `=` or the next argument, since no boolean flag takes a separate one.
func childArguments(arguments []string, lintConfigFileName string) []string {
	if lintConfigFileName == "" {
		return arguments
	}
	kept := make([]string, 0, len(arguments)+2)
	for index := 0; index < len(arguments); index++ {
		name, _, joined := strings.Cut(strings.TrimLeft(arguments[index], "-"), "=")
		if name == "lint-config" && strings.HasPrefix(arguments[index], "-") {
			if !joined {
				index++
			}
			continue
		}
		kept = append(kept, arguments[index])
	}
	return append(kept, "--lint-config", lintConfigFileName)
}

// isTheLocatedProject reports whether discovery found exactly the one project the walk up located, which
// a run then checks in this process, the way it always did.
func (found discovery) isTheLocatedProject(location projectLocation, locateError error) bool {
	if locateError != nil || len(found.Projects) != 1 {
		return false
	}
	project := found.Projects[0]
	return project.Engine == location.Engine && project.ConfigFile == "" &&
		filepath.Clean(filepath.Join(found.Root, project.Directory)) == filepath.Clean(location.Root)
}

// notes are the lines a run carries about markers discovery found and did not run, which a reader could have
// expected checked: those the root's ignorePatterns refuse, solution roots, and projects left with no file of
// their own. Ignored and dependency directories are not named, since skipping them is what everyone
// expects, and neither is a nested repository: one inside a program, such as Structure or Base, is checked
// as part of it, its drift reported read-only, as it always was. They are counted on the summary line.
func (found discovery) notes() []string {
	var lines []string
	for _, marker := range found.Refused {
		lines = append(lines, fmt.Sprintf("%s not checked: the root's ignorePatterns match it", marker))
	}
	for _, label := range found.Ownership.Solutions {
		lines = append(lines, fmt.Sprintf("%s not run: a solution tsconfig, which includes no file and references the projects checked instead", label))
	}
	for _, label := range found.Ownership.Yielding {
		lines = append(lines, fmt.Sprintf("%s not run: every file it includes belongs to a nearer tsconfig", label))
	}
	for label := range found.Ownership.SharedDirectory {
		lines = append(lines, fmt.Sprintf("%s ran with the cache off: another tsconfig's run keeps its cache in the same directory", label))
	}
	sort.Strings(lines[len(lines)-len(found.Ownership.SharedDirectory):])
	return lines
}

// projectRun is one project's run, as its child reported it.
type projectRun struct {
	project  discoveredProject
	output   []byte
	exitCode int
	err      error
	elapsed  time.Duration
	// notChecked, when set, is why the project was not run at all: a gap the report names, never a pass.
	notChecked string
}

// swiftFormatOnlyGap is why a Swift project is not run under --format-only. The engine has no format-only
// mode, so a run would type-check and lint the package as well: a format gate on cohere's own checkout
// spent 2m38s compiling swift/ and its tests, and at the house's load ran past the caller's patience
// (#nqb3mjv). Skipped and named until the engine's mode lands, so a mixed run still checks its other
// projects. A run whose every project is skipped checked nothing, and fails (#71a0ts8).
const swiftFormatOnlyGap = "not format-checked: the Swift engine has no format-only mode yet, and running it would " +
	"type-check and lint the package as well (#nqb3mjv)"

// runProjects checks every discovered project, each in its own run of this binary, all at once, and
// reports them as one: a section per project in the order discovery found them, then one summary line.
// It exits with the worst exit code among them, so the run is green only when every project is, and a
// project whose run crashed or could not start fails it.
//
// Each project is a child process rather than a goroutine, because a run is a whole program: its own
// type graph, its own settings and cache table, and its own memory ceiling. A child is exactly the run
// `cohere --directory <project>` makes alone, so a project's section reads the same as its own run's, and
// its exit code is the one that run alone would give, less what it yields to nearer projects.
//
// At most projectRunsAtOnce run together. Each builds a whole type graph and checks it on every core, so
// TanStack Query's 95 projects started at once were 95 graphs in memory, and more children than that buys no
// time the cores do not already give one.
func runProjects(found discovery, arguments []string, out io.Writer) int {
	executable, err := os.Executable()
	if err != nil {
		fmt.Fprintf(out, "cohere: finding this binary to run each project: %v\n", err)
		return 1
	}
	yieldDirectory, err := os.MkdirTemp("", "cohere-yield-")
	if err != nil {
		fmt.Fprintf(out, "cohere: making a place for what each project yields: %v\n", err)
		return 1
	}
	defer os.RemoveAll(yieldDirectory)

	children := forwardSignalsToChildren()
	defer children.stop()

	runs := make([]projectRun, len(found.Projects))
	var group sync.WaitGroup
	running := make(chan struct{}, projectRunsAtOnce())
	formatOnly := flagValue("format-only") == "true"
	for index, project := range found.Projects {
		if formatOnly && project.Engine == engineSwift {
			runs[index] = projectRun{project: project, notChecked: swiftFormatOnlyGap}
			continue
		}
		label := projectLabel(project)
		yieldFile, err := writeProjectYield(found.Ownership.Yields[label], yieldDirectory)
		if err != nil {
			runs[index] = projectRun{project: project, exitCode: 1, err: fmt.Errorf("writing what it yields: %w", err)}
			continue
		}
		projectArguments := arguments
		// TypeScript runs alone: the cache it turns off is cohere's table in the directory, which a Swift run
		// never reads or writes, and the Swift engine refuses --no-cache by name.
		if found.Ownership.SharedDirectory[label] && project.Engine == engineTypeScript {
			projectArguments = append(slices.Clone(arguments), "--no-cache")
		}
		group.Add(1)
		go func() {
			defer group.Done()
			running <- struct{}{}
			defer func() { <-running }()
			runs[index] = runProject(executable, found.Root, project, projectArguments, yieldFile, children)
		}()
	}
	group.Wait()

	// In the default view a project's own run ends on its footer, which leads with its path, so the projects
	// follow one another with no heading between them; `--verbose` keeps the sections and their account.
	account := accountOutput(out)
	worst := 0
	for _, checked := range runs {
		fmt.Fprintf(account, "%s\n", sectionHeading(checked.project))
		if checked.notChecked != "" {
			fmt.Fprintf(account, "%s\n\n", checked.notChecked)
			// Under --json the overall footer, which names every gap, is not printed, so the gap goes where a
			// person running the program still sees it.
			if activeOutput.Mode == outputJSON {
				fmt.Fprintf(os.Stderr, "cohere: %s %s\n", projectLabel(checked.project), checked.notChecked)
			}
			continue
		}
		out.Write(checked.output)
		if len(checked.output) > 0 && checked.output[len(checked.output)-1] != '\n' {
			fmt.Fprintln(out)
		}
		if checked.err != nil {
			fmt.Fprintf(out, "cohere: the %s run did not finish: %v\n", checked.project.Engine, checked.err)
		}
		fmt.Fprintln(account)
		if checked.exitCode > worst {
			worst = checked.exitCode
		}
	}

	fmt.Fprintf(account, "projects under %s:\n", found.Root)
	for _, checked := range runs {
		if checked.notChecked != "" {
			fmt.Fprintf(account, "  %s (%s): %s\n", projectLabel(checked.project), checked.project.Engine, checked.notChecked)
			continue
		}
		verdict := "green"
		if checked.exitCode != 0 {
			verdict = fmt.Sprintf("failed, exit %d", checked.exitCode)
		}
		fmt.Fprintf(account, "  %s (%s): %s in %s\n", projectLabel(checked.project), checked.project.Engine, verdict, round(checked.elapsed))
	}
	for _, note := range found.notes() {
		fmt.Fprintf(account, "  %s\n", note)
	}
	fmt.Fprintln(account, summaryLine(found, runs))
	// The overall verdict, last, in the human views. Under `--json` each project's own summary line, with its
	// label, is the record, and a program reads the worst from them.
	facts := overallFactsOf(runs, time.Since(processStart))
	facts.Root = found.Root
	if activeOutput.Mode != outputJSON {
		fmt.Fprintln(out, overallFooter(facts, activeOutput.Style))
	}
	// A run whose every project was skipped checked nothing, and fails the way a run that found none does
	// (#71a0ts8). Under --json the exit code is what says so, beside each gap on stderr.
	if len(facts.NotChecked) == len(runs) {
		worst = max(worst, 1)
		if activeOutput.Mode == outputJSON {
			fmt.Fprintf(os.Stderr, "cohere: nothing checked under %s: every project found was skipped\n", found.Root)
		}
	}
	// A run stopped by a signal ends the way the signal would have ended it, after every child has, so the
	// caller that sent it reads it as stopped and not as a verdict.
	if received := children.received(); received != 0 {
		return 128 + received
	}
	return worst
}

// overallFactsOf is what the overall verdict says of the projects' runs: how many of each engine, the ones
// that failed with their exits, and the ones whose run did not finish, which checked nothing.
func overallFactsOf(runs []projectRun, total time.Duration) overallFacts {
	facts := overallFacts{Total: total, ProjectsByEngine: map[string]int{}}
	for _, checked := range runs {
		if checked.notChecked != "" {
			facts.NotChecked = append(facts.NotChecked, projectLabel(checked.project)+" "+checked.notChecked)
			continue
		}
		facts.ProjectsByEngine[string(checked.project.Engine)]++
		switch {
		case checked.err != nil:
			facts.Unfinished++
		case checked.exitCode != 0:
			facts.Failed = append(facts.Failed, fmt.Sprintf("%s (exit %d)", projectLabel(checked.project), checked.exitCode))
		}
	}
	return facts
}

// projectRunsAtOnce is how many projects' runs share the machine: a quarter of its cores, and at least one.
// Each run checks its own program on every core already, so the bound costs little time, and it is what
// keeps a laptop alive on a repository with a hundred tsconfigs.
func projectRunsAtOnce() int {
	return max(1, runtime.NumCPU()/4)
}

// runProject runs one project's check as a child and keeps everything it printed. yieldFile, when set,
// names what it yields to nearer projects (see ownership.go). The child is held in children while it runs,
// so a signal to this run reaches it.
func runProject(executable string, root string, project discoveredProject, arguments []string, yieldFile string, children *projectChildren) projectRun {
	directory := filepath.Join(root, filepath.FromSlash(project.Directory))
	prefix := []string{"--directory", directory}
	if project.ConfigFile != "" {
		prefix = append(prefix, "--tsconfig", project.ConfigFile)
	}
	child := exec.Command(executable, append(prefix, arguments...)...)
	// Started in the project, so its report does not add that it checked somewhere other than where it
	// started: the section's heading already says which project it is.
	child.Dir = directory
	// The verdict descriptor is this run's: a child's verdict reaches this process as its exit code, and
	// only the merged one may answer the caller.
	child.Env = append(withoutVariable(withoutVariable(os.Environ(), VerdictVariable), projectYieldVariable),
		projectEngineVariable+"="+string(project.Engine), projectLabelVariable+"="+projectLabel(project))
	if yieldFile != "" {
		child.Env = append(child.Env, projectYieldVariable+"="+yieldFile)
	}
	var output bytes.Buffer
	child.Stdout = &output
	child.Stderr = &output

	started := time.Now()
	err := children.run(child)
	run := projectRun{project: project, output: output.Bytes(), elapsed: time.Since(started)}
	var exited *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &exited) && exited.ExitCode() > 0:
		run.exitCode = exited.ExitCode()
	default:
		// Killed by a signal, or never started: a project that was not checked is a failure.
		run.exitCode = 1
		run.err = err
	}
	return run
}

// withoutVariable is an environment with one variable removed.
func withoutVariable(environment []string, name string) []string {
	kept := make([]string, 0, len(environment))
	for _, entry := range environment {
		if !strings.HasPrefix(entry, name+"=") {
			kept = append(kept, entry)
		}
	}
	return kept
}

// projectLabel names a project by its directory, the root by ".", or by its tsconfig's path when that is
// not the directory's tsconfig.json.
func projectLabel(project discoveredProject) string {
	if project.ConfigFile != "" {
		return path.Join(project.Directory, project.ConfigFile)
	}
	return project.Directory
}

// sectionHeading opens a project's section.
func sectionHeading(project discoveredProject) string {
	return fmt.Sprintf("== %s (%s) ==", projectLabel(project), project.Engine)
}

// summaryLine closes the report: how many projects were checked, of which engines, how many failed, and
// what the walk did not enter.
func summaryLine(found discovery, runs []projectRun) string {
	counts := map[projectEngine]int{}
	failed := 0
	checkedCount := 0
	notChecked := 0
	for _, checked := range runs {
		if checked.notChecked != "" {
			notChecked++
			continue
		}
		checkedCount++
		counts[checked.project.Engine]++
		if checked.exitCode != 0 {
			failed++
		}
	}
	engines := []string{}
	for _, engine := range []projectEngine{engineTypeScript, engineSwift} {
		if counts[engine] > 0 {
			engines = append(engines, fmt.Sprintf("%d %s", counts[engine], engine))
		}
	}
	verdict := "every one green"
	if failed > 0 {
		verdict = fmt.Sprintf("%d failed", failed)
	}
	notRun := ""
	if len(found.Ownership.Solutions) > 0 || len(found.Ownership.Yielding) > 0 {
		notRun = fmt.Sprintf("; not run: %d solution tsconfigs, %d whose every file a nearer tsconfig owns",
			len(found.Ownership.Solutions), len(found.Ownership.Yielding))
	}
	if notChecked > 0 {
		notRun += fmt.Sprintf("; not checked: %d, each named above", notChecked)
	}
	return fmt.Sprintf("projects: %d checked (%s), %s%s; not searched for projects: %d directories .gitignore ignores, "+
		"%d dependency, build, cache or fixture directories, %d nested repositories",
		checkedCount, strings.Join(engines, ", "), verdict, notRun, found.Ignored, found.NeverDescended,
		len(found.NestedRepositories)+found.Submodules)
}

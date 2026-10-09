package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/suppression"
	"github.com/system-inc/cohere/internal/types/program"
	"github.com/system-inc/cohere/internal/types/sourcename"
)

// applyProposedFixes rewrites every file whose rules proposed a repair, and reports what it did.
//
// Each file converges on its own rather than the tree converging as a whole. A fix is local to the
// file it edits, so a per-file fixpoint reaches the same result as a whole-tree loop while re-linting
// only the files that actually changed — and it means one pathological file cannot force every other
// file through another pass.
//
// Only files with at least one proposal are considered. A file nobody proposed anything for is never
// opened, never re-parsed, and never written, which is what keeps a fix run proportional to the work
// rather than to the tree.
//
// With write false this is `--no-fix`: every file goes through the same fixpoint and transform and
// nothing reaches the disk, so the summary is what a writing run would do, stated as what would
// change. One function with a switch rather than a second copy, because a check that decided
// anything differently from the write would report a tree as clean that `--fix` then rewrites.
func applyProposedFixes(
	ctx context.Context,
	graph *program.Graph,
	projectFiles []*ast.SourceFile,
	rules []rule.Rule,
	formatPlan func() (edit.Transform, []string),
	writable formatScope,
	repositoryRoot string,
	maxPasses int,
	write bool,
	early *earlyFormat,
) (edit.Summary, program.Result, *crlfFiles, error) {
	// This walk is the run's FIRST walk over the program, and it is returned so the lint phase can
	// reuse it rather than repeat it.
	//
	// An earlier version of this comment said the opposite: that a reporting walk had already
	// happened and this one merely asked again, cheaply. Both halves were wrong. Fix is phase 2 and
	// lint runs after phase 3, so nothing walks before this; and measured on the ahra tree a run
	// costs roughly a fixed 0.9s plus 1.1s per walk, so this is about a third of a default run
	// rather than a cheap repeat. The judgment was disclosed at the call site, which is the only
	// reason it was checkable at all.
	//
	// Returning the result rather than dropping it is what lets the caller decide. That decision is
	// the caller's and not this function's, because only the caller knows whether anything was
	// rewritten afterward: a file rewritten here makes these diagnostics describe bytes that no
	// longer exist.
	//
	// No program file in scope is a run about files the program does not hold, such as
	// `cohere --fix --format notes.md`, and there are no rules to ask. Walking anyway refused the empty
	// set ("nothing to walk"), which bailed the phase before the format candidates below were ever
	// formatted, so naming a .md, .css, .json or .graphql formatted nothing. The empty result is not
	// handed on as a walk: the caller reuses it only when projectFiles is non-empty.
	// Counted on the way into the formatter, so a checkout git wrote with CRLF is named once below rather
	// than left to read as a tree of unformatted files.
	lineEndings := &crlfFiles{}
	if early != nil {
		lineEndings = early.lineEndings
	}
	// The transform and the candidates come from the format scope. Without an early pass they are needed now, to
	// start the speculation. With one, the speculation is already running from the early pass's own scope, and the
	// walk needs neither, so they are taken only once the walk is done: on a cold ahra run the scope was ready
	// 36 to 51ms after the graph, and the walk used to wait for it (#g3046x5).
	var transform edit.Transform
	var formatCandidates []string
	planTaken := false
	takePlan := func() {
		if planTaken {
			return
		}
		planTaken = true
		transform, formatCandidates = formatPlan()
		if transform != nil {
			transform = lineEndings.observing(transform)
		}
	}
	if early == nil {
		takePlan()
	}

	// The format pass starts now, beside the walk and the type check, rather than after them (#679s763).
	// Every format candidate is formatted as if no rule proposed anything for it; after the walk, a result
	// is kept only for a file the walk proposed nothing for and whose bytes are the ones the walk read,
	// and every other file takes the path it always took. See speculateFormat.
	var speculation *formatSpeculation
	if early != nil {
		// Begun during the graph build on every idle core; the walk wants the cores now, so it keeps the share it
		// always had. See startEarlyFormat.
		early.narrow(speculationWorkers())
	} else {
		var programs programOffer
		if graph != nil {
			programs.offer(graph)
		}
		speculation = speculateFormat(formatCandidates, transform, maxPasses, &programs)
	}

	var result program.Result
	if len(projectFiles) > 0 {
		walked, err := graph.Walk(ctx, projectFiles, rules)
		if err != nil {
			if early != nil {
				speculation = early.speculationOf()
			}
			speculation.wait()
			return edit.Summary{}, program.Result{}, nil, fmt.Errorf("collecting proposals: %w", err)
		}
		// Before anything the walk prints, so the scope's own notes come where they always came.
		takePlan()
		reportForeignCheckers(graph, walked, len(projectFiles))
		result = walked
	}
	takePlan()
	if early != nil {
		speculation = early.speculationOf()
	}
	if graph != nil && graph.FusedCheck != nil {
		speculation.stop()
	}

	// Group proposals by the file they belong to. A diagnostic carries its source file, so the
	// grouping is exact rather than inferred from a range.
	//
	// A proposal in a file outside the write scope is withheld here, before the file can become a
	// candidate. projectFiles is what this run checks, and that set grows past what the caller stated:
	// importers of a named file, or the whole tree past the closure limit. Checking them is honest;
	// rewriting them is not, because nobody asked for those files to change. The finding still
	// reaches the lint phase, so a withheld repair is reported rather than lost.
	//
	// A whole-tree run has no stated set to filter against, so it is bounded the way the format walk is:
	// it writes only repositoryRoot, the repository the run is in. Nothing is written inside a repository
	// of its own below it (a submodule, or any directory with its own `.git`), nor, for a run started
	// inside a library, in the project outside it. The files are still checked and their findings still
	// reported; only the write is withheld, and counted per repository so the run says what it did not
	// apply. A run that names a path has asked for it, and its stated scope admits the file as before.
	byFileName := map[string][]edit.Proposal{}
	withheld := map[string]struct{}{}
	inNestedRepository := map[string]int{}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.SourceFile == nil || len(diagnostic.Fixes) == 0 {
			continue
		}
		// In the form the format scope holds, since both halves of the union below are keyed by it. The
		// compiler names a file `C:/work/a.ts` and the walk `C:\work\a.ts` on Windows, and one file under
		// both names was fixed under one and formatted under the other, from the text before its fixes.
		// Clean converts the separators there and changes nothing elsewhere.
		fileName := filepath.Clean(diagnostic.SourceFile.FileName().AsString())
		if !writable.Everything && !writable.includes(fileName) {
			withheld[fileName] = struct{}{}
			continue
		}
		if writable.Everything && repositoryRoot != "" {
			if elsewhere := unwritableRepository(repositoryRoot, fileName); elsewhere != "" {
				inNestedRepository[elsewhere] += len(diagnostic.Fixes)
				continue
			}
		}
		for _, proposed := range diagnostic.Fixes {
			byFileName[fileName] = append(byFileName[fileName], edit.Proposal{
				RuleName: diagnostic.RuleName,
				Fix:      proposed,
			})
		}
	}
	reportWithheld(withheld, writable)
	reportInNestedRepositories(accountOutput(os.Stderr), inNestedRepository)

	// The file set is the union of "something proposed a fix here" and "this file can be formatted",
	// and the union rather than the intersection is the whole point. Fixing is driven by findings, so
	// it only visits files a rule complained about. Formatting is not: a file can be perfectly correct
	// and badly formatted, and a formatter that only ran where a rule had already fired would never
	// touch most of a tree while reporting a clean run.
	//
	// When no transform is configured the union collapses back to the files with proposals, so a
	// fix-only run costs exactly what it did before.
	candidates := map[string]struct{}{}
	for fileName := range byFileName {
		candidates[fileName] = struct{}{}
	}
	// The formattable half comes from the scope rather than from the program, and this is the second
	// half of the split the ruling made. A proposed fix comes from a rule that ran over the program,
	// so its candidate must be in the program. A format candidate comes from the disk.
	//
	// Taking them from projectFiles is what made css, markdown, json and yaml invisible: a tsconfig
	// enumerates TypeScript by construction, so a `.css` was never a candidate, and the coverage line
	// could not even report it as skipped. The scope already holds the right set, walked from the
	// project root with the ignore layers applied and filtered by what the engine handles.
	//
	// These are not filtered against the write scope, because the caller already built the format
	// scope from it: a run with stated paths formats exactly those paths. A second filter here would
	// answer the same question twice and make neither answer observable.
	for _, fileName := range formatCandidates {
		candidates[fileName] = struct{}{}
	}

	if len(candidates) == 0 {
		speculation.wait()
		// An empty run still reports its population, so "nothing proposed a fix" cannot be confused
		// with "the fixer never ran".
		summary := edit.Summarize(nil)
		summary.Checked = !write
		return summary, result, lineEndings, nil
	}

	process := edit.FixAndTransformFile
	if !write {
		process = edit.CheckFile
	}

	// Sorted so a run is reproducible and a diff of two runs is readable.
	fileNames := make([]string, 0, len(candidates))
	for fileName := range candidates {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	// linted is every file the walk ran the rules over, in the form the candidates are keyed by.
	linted := make(map[string]bool, len(projectFiles))
	for _, sourceFile := range projectFiles {
		linted[filepath.Clean(sourceFile.FileName().AsString())] = true
	}

	// propose is what a file's fixpoint asks for proposals. The first pass reuses the proposals already
	// collected; later passes re-lint the rewritten text. Reusing them for the first pass and only the
	// first pass is what keeps offsets honest: a proposal is valid exactly against the text it was
	// computed from.
	//
	// relint is how a later pass gets its proposals: the rules on the serial path, or, on the parallel
	// one, a refusal that sends the file back to the serial path.
	propose := func(fileName string, relint edit.Propose) edit.Propose {
		firstPass := byFileName[fileName]
		used := false

		// A format candidate inside a nested repository is formatted and never fixed on a whole-tree
		// run. Its first-pass proposals were withheld above, and without this its later passes re-linted
		// the formatted text and applied them anyway: formatting one file in nexus rewrote it under
		// nexus/consistency-no-multiline-arrow-function on a run that reported the repository's fixes as
		// not applied.
		//
		// A format candidate the walk never linted is formatted and never fixed either, for the same reason:
		// no rule ran over it, so no finding was reported for it, and a fix landing there is a repair nobody
		// was told about. It was costly as well as wrong. angular/angular's deploy-docs-site action keeps a
		// 2.8 MB esbuild bundle beside a tsconfig that does not include it; the format pass printed it, every
		// later pass re-linted all 2.8 MB with every rule, and `--no-fix` reported prefer-template, strict and
		// more as rewrites of a vendored file the lint phase never looked at (#xn1k1gz).
		fixesWithheld := !linted[fileName] ||
			writable.Everything && repositoryRoot != "" && unwritableRepository(repositoryRoot, fileName) != ""

		return func(_ string, text string) ([]edit.Proposal, error) {
			if !used {
				used = true
				return firstPass, nil
			}
			if fixesWithheld {
				return nil, nil
			}
			return relint(fileName, text)
		}
	}
	serially := func(fileName string, text string) ([]edit.Proposal, error) {
		return proposalsForText(fileName, text, graph, rules)
	}

	// A speculative result stands for a file only where the parallel pass below would have computed the
	// same one: no proposal, and the bytes the walk read. The rest are formatted here, as before.
	speculated := speculation.keepable(byFileName, graph)
	attempts := formatInParallel(fileNames, byFileName, speculated, func(fileName string) (edit.FileResult, error) {
		return watchedFile(os.Stderr, slowFixNotice, fileName, func() (edit.FileResult, error) {
			return process(fileName, propose(fileName, refuseToRelint), transform, maxPasses)
		})
	})

	results := make([]edit.FileResult, 0, len(fileNames))
	for index, fileName := range fileNames {
		fileResult, err := attempts[index].result, attempts[index].err
		if !attempts[index].done {
			fileResult, err = watchedFile(os.Stderr, slowFixNotice, fileName, func() (edit.FileResult, error) {
				return process(fileName, propose(fileName, serially), transform, maxPasses)
			})
		}
		if err != nil {
			// One file failing must not abandon the rest. The failure is reported rather than
			// swallowed, and the tree is left in a state where every other fix still landed.
			fmt.Fprintf(os.Stderr, "cohere: %v\n", err)

			// A failure is its own category rather than a flavor of "did not converge". The two send a
			// reader somewhere different: not-converged means the pass budget ran out with work still
			// landing, which points at two rules arguing, while a failure means the file never entered
			// the loop at all. A zero-value FileResult has Converged false, so leaving it alone would
			// file this under the wrong heading — a real number reported as the wrong finding.
			fileResult.FileName = fileName
			fileResult.Failed = true
			fileResult.Changed = false
		}
		results = append(results, fileResult)
	}

	lineEndings.report(accountOutput(os.Stderr), write)

	summary := edit.Summarize(results)
	summary.Checked = !write
	return summary, result, lineEndings, nil
}

// slowFixNotice is how long one file may spend in the fix phase before the run names it. A healthy file takes
// milliseconds and the 2.8 MB bundle that found this takes about two seconds now; a file past this is a runaway
// the run would otherwise sit in with nothing on screen, as angular/angular's deploy-docs-site bundle did for
// 26 minutes and 19 GB (#xn1k1gz).
const slowFixNotice = 30 * time.Second

// watchedFile runs one file's fix and format, and names the file on out if it is still running after after.
// The run goes on: the notice says which file to report, or to leave out, while it is still happening rather
// than after a wait nobody can explain.
func watchedFile(out io.Writer, after time.Duration, fileName string, process func() (edit.FileResult, error)) (edit.FileResult, error) {
	notice := time.AfterFunc(after, func() {
		fmt.Fprintf(out, "cohere: %s is still being fixed and formatted after %s, which no file should take. This is "+
			"likely a bug in cohere, not in your code: please report it at %s with the output of `cohere --version` "+
			"and the file.\n", fileName, after, reportBugsAt)
	})
	defer notice.Stop()
	return process()
}

// errRelintRefused is what the parallel format pass's proposer answers when a file would need its
// rules run again, which sends the file back to the serial path.
var errRelintRefused = errors.New("re-linting is left to the serial path")

// refuseToRelint is the parallel pass's answer to a later fix pass: no rule runs off the serial path.
func refuseToRelint(string, string) ([]edit.Proposal, error) {
	return nil, errRelintRefused
}

// parallelFormatWorkers is how many files the parallel pass formats at once. A variable so a test can
// set it to zero, which leaves every file to the serial path, and compare the two runs.
var parallelFormatWorkers = func() int { return runtime.GOMAXPROCS(0) }

// formatAttempt is one file's result from the parallel pass. done is false when the file is left to the
// serial path, because it had proposals or its formatting led back to the rules.
type formatAttempt struct {
	result edit.FileResult
	err    error
	done   bool
}

// formatInParallel runs the fix phase's own processing, in parallel, over the candidates no rule proposed
// anything for, and returns one attempt per file name, in the same order.
//
// A first run with no format record visits every file the formatter handles: 5,090 on ahra, one at a
// time, took 84.6s. Almost all of them have no proposals and are already formatted, and for those the
// whole of the work is a parse and a print, both pure per file. So they are done here at once.
//
// The rules are another matter: a rule may keep state of its own, and nothing has shown that every one
// of them can run beside itself. So nothing here runs one. A file whose formatting changed its text
// would be re-linted next, and its proposer refuses instead (refuseToRelint), which discards the attempt
// before anything is written and leaves the file to the serial loop. That loop then does exactly what it
// always did, in the same order, so the run's findings, its summary and what it writes are the serial
// run's, and only the already-formatted majority moved.
//
// speculated holds what speculateFormat already computed for some of those files, which is taken as their
// attempt rather than computed again.
func formatInParallel(fileNames []string, byFileName map[string][]edit.Proposal, speculated map[string]formatAttempt,
	process func(string) (edit.FileResult, error)) []formatAttempt {
	attempts := make([]formatAttempt, len(fileNames))
	next := make(chan int, len(fileNames))
	for index, fileName := range fileNames {
		if len(byFileName[fileName]) != 0 {
			continue
		}
		if attempt, found := speculated[fileName]; found {
			attempts[index] = attempt
			continue
		}
		next <- index
	}
	close(next)

	var workers sync.WaitGroup
	for range parallelFormatWorkers() {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range next {
				result, err := process(fileNames[index])
				if errors.Is(err, errRelintRefused) {
					continue
				}
				attempts[index] = formatAttempt{result: result, err: err, done: true}
			}
		}()
	}
	workers.Wait()
	return attempts
}

// printWouldChange reports each file a `--no-fix` run found that `--fix` would rewrite, one finding
// per file, in the shape every other finding prints so an editor's problem matcher and a reader keyed
// on `[rule/id]` both see it.
//
// A file, not a fix, is the unit: formatting rewrites a file as a whole, and the fixable findings
// behind a repair already print under lint at their own positions. What this adds is the fact the
// rest of the run cannot state, that this file is not what the gate would leave, and which rules or
// the formatter would change it. Before it, an unformatted file passed a clean `--no-fix` run unseen.
//
// A file that arrived with CRLF says so, with the .gitattributes line that stops git writing it that way:
// the run's CRLF note prints only under --verbose, and without the cause a fresh Windows checkout reads
// as a tree of unformatted files (#dr78rt8).
func printWouldChange(out io.Writer, changed []edit.ChangedFile, lineEndings *crlfFiles) {
	for _, file := range changed {
		message := "--fix would rewrite this file: " + strings.Join(file.Changers, ", ")
		if lineEndings.has(file.FileName) {
			message += crlfReason
		}
		printFinding(out,
			runFinding{Path: file.FileName, Line: 1, Column: 1, Severity: "error", Rule: "fix", MessageID: "would-change", Message: message},
			fmt.Sprintf("%s:1:1 - %s [fix/would-change]\n", file.FileName, message))
	}
}

// reportWithheld names the files whose repairs were withheld because they sit outside what the
// caller stated.
//
// Named rather than counted, and capped so a whole-tree fallback over thousands of files stays one
// readable note. Silence here would let a reader believe the run repaired everything it could, when
// the files it checked beyond the stated scope still carry fixable findings.
func reportWithheld(withheld map[string]struct{}, writable formatScope) {
	if len(withheld) == 0 {
		return
	}
	names := make([]string, 0, len(withheld))
	for fileName := range withheld {
		names = append(names, fileName)
	}
	sort.Strings(names)

	const shown = 5
	listed := names
	if len(listed) > shown {
		listed = listed[:shown]
	}
	more := ""
	if len(names) > shown {
		more = fmt.Sprintf(" and %d more", len(names)-shown)
	}

	stated := writable.RequestDescription
	if stated == "" {
		stated = writable.Description
	}
	verb := "were"
	if len(names) == 1 {
		verb = "was"
	}
	fmt.Fprintf(accountOutput(os.Stderr),
		"note: %d file%s outside %s had fixable findings and %s not rewritten, because only what was stated is written: %s%s\n",
		len(names), plural(len(names)), stated, verb, strings.Join(listed, ", "), more)
}

// reportInNestedRepositories says, per repository, how many fixes a whole-tree run did not apply
// because they sit outside the repository the run writes: in a repository of its own below it, or in the
// project outside a library the run started in.
//
// Said every time there are any, because the alternative is the defect this replaces in reverse: a
// run that quietly leaves fixable findings in place reads as a run that fixed everything it could.
func reportInNestedRepositories(out io.Writer, fixesByRepository map[string]int) {
	repositories := make([]string, 0, len(fixesByRepository))
	for repository := range fixesByRepository {
		repositories = append(repositories, repository)
	}
	sort.Strings(repositories)
	for _, repository := range repositories {
		count := fixesByRepository[repository]
		noun := "fixes"
		if count == 1 {
			noun = "fix"
		}
		fmt.Fprintf(out, "note: %d %s not applied: in %s, which a run here does not write unless it is named\n",
			count, noun, repository)
	}
}

// proposalsForText re-runs the rules against rewritten text.
//
// The file is re-parsed rather than re-read, because the text in hand is the candidate that has not
// been written yet. This pass has no type information: a re-parse produces a standalone file rather
// than a member of the program, and building a whole new program per pass would cost seconds. A
// type-aware rule that proposes a fix will therefore contribute on the first pass, where the real
// checker was present, and not on later ones. That is a real limitation and it is stated rather than
// hidden — the alternative is a fix run that rebuilds a 9,530-file program once per pass per file,
// which is precisely the cost this tool exists to remove.
func proposalsForText(
	fileName string,
	text string,
	graph *program.Graph,
	rules []rule.Rule,
) ([]edit.Proposal, error) {
	// No rules propose nothing, and need no graph to say so: `--format-only` has neither.
	if len(rules) == 0 {
		return nil, nil
	}
	rooted := fileName
	if !tspath.IsRootedDiskPath(rooted) {
		rooted = "/" + rooted
	}
	rooted = tspath.NormalizePath(rooted)

	// The kind the program parsed the file as: an Adamic `.a` name is read as `.ts`, as the program reads it
	// (#6mhafvb). Unknown, which the bare name gives, is a parser panic.
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: tspath.RootedFilePath(rooted),
		PathKey:  tspath.PathKey(rooted),
	}, text, core.GetScriptKindFromFileName(tspath.RootedFilePath(sourcename.TreatedAs(rooted))))
	if sourceFile == nil {
		return nil, fmt.Errorf("re-parsing %s produced nothing", fileName)
	}

	/*
	 * Options are resolved here, not left nil, and the omission was a real defect rather than a
	 * simplification.
	 *
	 * A rule handed nil options falls back to whatever it does unconfigured, and for several rules
	 * that is to decline entirely: `import-require-path-alias` reads its alias map from options, so
	 * with nil it proposes nothing. That made every pass after the first blind to it. The visible
	 * symptom was a file needing two separate `--fix` invocations to settle, because the first pass
	 * applied one fix and the second pass could no longer see the other one to apply it.
	 *
	 * It hid because the first pass reuses the proposals the lint phase already collected, which do
	 * have options. Only passes two and later go through here, and only a file with two overlapping
	 * fixes ever needs a second pass at all.
	 */
	applicable := rules
	ruleOptions := map[string]any{}
	if graph.LintConfig != nil {
		resolution := graph.LintConfig.Resolve(fileName)
		if resolution.Ignored {
			return nil, nil
		}
		applicable = applicable[:0:0]
		for _, subject := range rules {
			if !resolution.Enabled(subject.Name) {
				continue
			}
			decoded, decodeError := graph.RuleOptions.Decode(subject.Name, resolution.RawOptionsFor(subject.Name))
			if decodeError != nil {
				return nil, fmt.Errorf("decoding options for %s on %s: %w", subject.Name, fileName, decodeError)
			}
			ruleOptions[subject.Name] = decoded
			applicable = append(applicable, subject)
		}
	}

	/*
	 * A rule that needs the type checker cannot run here, and running it anyway crashed the whole
	 * command.
	 *
	 * This pass re-lints text that has just been rewritten, which by definition is not the text the
	 * program was built from, so no checker in the graph describes it: `ctx.TypeChecker` is nil and
	 * a rule declaring `NeedsTypeChecker` dereferences it on its first question.
	 * `react/no-danger-with-children` is the one that found it, at `resolvedDeclaration`.
	 *
	 * It stayed hidden because reaching this code takes two things at once: a fix that lands, and a
	 * second pass to check it. Every fixable rule in the tree happened to be syntactic, so until a
	 * fixer landed on a file that also carries a checker-needing rule, the pass simply never ran
	 * with one in the list.
	 *
	 * Skipping is the honest answer rather than a workaround. A type-aware rule has nothing true to
	 * say about text no program contains, so its verdict here would be a guess whichever way it
	 * came out. The cost is bounded and worth naming: a fix that introduces a type-aware violation
	 * is not caught by the pass that applied it, and is caught by the next real lint run.
	 */
	withoutTypeChecker := applicable[:0:0]
	for _, subject := range applicable {
		if subject.NeedsTypeChecker {
			continue
		}
		withoutTypeChecker = append(withoutTypeChecker, subject)
	}
	applicable = withoutTypeChecker

	/*
	 * The same suppression index the lint path builds, for the same reason.
	 *
	 * `internal/types/program/walk.go:500` filters a diagnostic through `directives.Suppresses`
	 * before reporting it. This path did not, so a rule's finding was correctly withheld and its
	 * FIX was applied anyway: both linters reported nothing on a disabled line and both fixers
	 * rewrote it. That is worse than either half alone, because the file changes with no diagnostic
	 * explaining why and the only evidence is in `git diff`.
	 *
	 * Demonstrated rather than theorised. `libraries/structure/source/router/hooks/useRouter.ts`
	 * carries a deliberate `eslint-disable-next-line nexus/import-no-forbidden-source`, and a fixer
	 * run reverted it, reintroducing a circular import worth 108 findings across five rules.
	 */
	directives := suppression.Build(sourceFile.Text())

	/*
	 * One file cache for every rule re-linting this text, as the lint walk keeps one per file.
	 *
	 * Without it rule.Cached recomputes on every ask, and a derivation meant to be paid once per file is
	 * paid wherever a rule asks: no-invalid-this asks for the file's comments at each function, so the
	 * whole-file comment scan ran once per function. On a 2.8 MB esbuild bundle in angular/angular, which
	 * the format pass printed and this path then re-linted, `--no-fix` took 26 minutes and 19 GB on one
	 * core while each half of it finished in seconds (#xn1k1gz).
	 */
	fileCache := rule.NewFileCache()

	var diagnostics []rule.Diagnostic
	for _, subject := range applicable {
		currentRule := subject
		context := rule.Context{
			SourceFile: sourceFile,
			Program:    rule.ViewProgram(graph.Program, sourceFile, currentRule),
			FileCache:  fileCache,
			Report: func(diagnostic rule.Diagnostic) {
				diagnostic.RuleName = currentRule.Name
				if diagnostic.SourceFile == nil {
					diagnostic.SourceFile = sourceFile
				}
				if directives.Suppresses(diagnostic.RuleName, diagnostic.Range.Pos()) {
					return
				}
				diagnostics = append(diagnostics, diagnostic)
			},
		}

		listeners := currentRule.Run(context, ruleOptions[currentRule.Name])
		if len(listeners) == 0 {
			continue
		}
		walkNode(sourceFile.AsNode(), listeners)
	}

	return edit.ProposalsFrom(diagnostics), nil
}

// walkNode visits every node, calling any listener registered for its kind.
func walkNode(node *ast.Node, listeners rule.Listeners) {
	if node == nil {
		return
	}
	if listener, isListening := listeners[node.Kind]; isListening {
		listener(node)
	}
	node.ForEachChild(func(child *ast.Node) bool {
		walkNode(child, listeners)
		return false
	})
}

// formatSpeculation is the format pass begun before the walk: one attempt per format candidate, computed as
// formatInParallel computes a file with no proposals, and kept or discarded once the walk is over.
type formatSpeculation struct {
	done     chan struct{}
	attempts map[string]formatAttempt
	// read is the text each attempt was computed from, so a result is kept only for those exact bytes.
	read map[string]string

	// stopped, once closed, has the workers take no more files: what they have not begun is left to
	// formatInParallel, which has every core. See stop.
	stopped  chan struct{}
	stopOnce sync.Once

	// limit is how many of the workers may take another file. See narrow.
	limit atomic.Int32
}

// speculateFormat starts formatting every candidate at once, on its own workers, and returns at once.
//
// The walk leaves cores idle: its workers wait on their checker's lock about half the time (#zqsdzbq), and
// the type check shares those checkers. Formatting needs no checker and no program, only each file's text
// and the options its directory resolves to, so it can fill that idle time rather than run after the walk.
// Measured on ahra (#679s763), the pass after the walk was 0.5 to 0.6s of a 3.2s cold run, and starting it
// early takes about 0.17s off: the rest is the type check, which then sets the run (speculationWorkers).
//
// It never writes. Each attempt goes through edit.CheckFile, and only a file its formatting left unchanged is
// kept, which is the common case: on a cold run nearly every file is already formatted. A file the printer
// changed is discarded, whether it came back refused (a TypeScript file, whose rules would be asked about the
// printed text) or simply changed (a markdown, css or json file, which has none), and is left to the path
// after the walk, which writes it. A file the walk then proposes a fix for, or whose bytes are not
// the ones the walk read, is discarded too (keepable), so what is kept is exactly what formatInParallel would
// have computed, and everything else is computed as it always was.
//
// A TypeScript file is formatted from the program's own tree where the program has one of exactly the bytes read,
// already bound, rather than parsed twice more, once by the fix engine's guard and once by the printer (#dk2502g).
// programs is where the program arrives once it is built, or nil. Every other file goes first, so the program
// has the time to be built while they are formatted, and a TypeScript file reached before then is parsed, as
// before: nothing waits for the program.
func speculateFormat(candidates []string, transform edit.Transform, maxPasses int, programs *programOffer) *formatSpeculation {
	return speculateFormatOn(candidates, transform, maxPasses, speculationWorkers(), programs)
}

// speculateFormatOn is speculateFormat on a given number of workers.
func speculateFormatOn(candidates []string, transform edit.Transform, maxPasses int, workers int, programs *programOffer) *formatSpeculation {
	speculation := &formatSpeculation{done: make(chan struct{}), attempts: map[string]formatAttempt{}, read: map[string]string{},
		stopped: make(chan struct{})}
	speculation.limit.Store(int32(workers))
	if transform == nil || len(candidates) == 0 || workers == 0 {
		close(speculation.done)
		return speculation
	}
	next := make(chan string, len(candidates))
	for _, fileName := range candidates {
		if !edit.TypeScriptParsable(fileName) {
			next <- fileName
		}
	}
	for _, fileName := range candidates {
		if edit.TypeScriptParsable(fileName) {
			next <- fileName
		}
	}
	close(next)
	var mutex sync.Mutex
	var group sync.WaitGroup
	for worker := range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for {
				select {
				case <-speculation.stopped:
					return
				default:
				}
				if int32(worker) >= speculation.limit.Load() {
					return
				}
				fileName, more := <-next
				if !more {
					return
				}
				first := true
				unproposed := func(_ string, _ string) ([]edit.Proposal, error) {
					if first {
						first = false
						return nil, nil
					}
					return refuseToRelint("", "")
				}
				var result edit.FileResult
				var err error
				if tree := programs.unchangedTreeOf(fileName); tree != nil {
					result, err = edit.CheckTextSeeded(fileName, tree.Text(), tree, unproposed, transform, maxPasses)
				} else {
					result, err = edit.CheckFileSeeded(fileName, programs.boundTreeOf(fileName), unproposed, transform, maxPasses)
				}
				// Only an unchanged result is final this early. A changed one is left to the path after the
				// walk, which writes it, or re-lints it where its type has rules: a markdown, css or json file
				// the printer changed comes back changed rather than refused, and keeping it would report a
				// rewrite that this check, which writes nothing, never made.
				if errors.Is(err, errRelintRefused) || err == nil && result.Changed {
					continue
				}
				mutex.Lock()
				speculation.attempts[fileName] = formatAttempt{result: result, err: err, done: true}
				speculation.read[fileName] = result.Text
				mutex.Unlock()
			}
		}()
	}
	go func() {
		group.Wait()
		close(speculation.done)
	}()
	return speculation
}

// programOffer is where a speculation finds the program once it is built. Its zero value has none, and so does
// a nil one.
type programOffer struct {
	program atomic.Pointer[compiler.Program]
	// graph is the graph the program was built in, set before the program, so a speculation that sees the
	// program can ask the graph whether its copy of a file is still the file's.
	graph atomic.Pointer[program.Graph]
}

// offer hands the speculation the graph's program, once it is built.
func (programs *programOffer) offer(graph *program.Graph) {
	if programs != nil && graph != nil && graph.Program != nil {
		programs.graph.Store(graph)
		programs.program.Store(graph.Program)
	}
}

// unchangedTreeOf is boundTreeOf's tree when the program's copy of the file is the file's bytes, by the stat it
// was read under and a file that starts with no byte order mark (Graph.ReadUnchanged): the speculation then formats the tree's own text and reads nothing (#q6dey77). Rereading every file to
// compare it with the tree was 51 MB on a cold ahra run. A file changed on disk since the program read it has
// another stat, so it is read as before, its bytes differ from the program's, and keepable discards the
// attempt: an edit made after the walk read the file is still noticed.
func (programs *programOffer) unchangedTreeOf(fileName string) *ast.SourceFile {
	tree := programs.boundTreeOf(fileName)
	if tree == nil || !programs.graph.Load().ReadUnchanged(fileName, tree.Text()) {
		return nil
	}
	return tree
}

// boundTreeOf is the program's tree of fileName if the program is in and has bound that file, or nil.
//
// Bound, because the binder writes node flags the formatter's converter reads (binder.go sets
// NodeFlagsThisNodeOrAnySubNodesHasError and NodeFlagsExportContext on nodes, and estree reads Let, Const,
// Using and Reparsed off the same word), and the checkers bind on their own goroutines while this pass runs.
// isBound is an atomic stored once the binder is done with the file, so a file that reads bound here has had
// every flag written before any is read. A file not yet bound is not waited for: it is parsed, as before.
//
// The tree is not yet known to be of the bytes the speculation reads; edit.CheckFileSeeded compares them.
func (programs *programOffer) boundTreeOf(fileName string) *ast.SourceFile {
	if programs == nil {
		return nil
	}
	program := programs.program.Load()
	if program == nil {
		return nil
	}
	sourceFile := program.GetSourceFile(tspath.RootedFilePath(fileName))
	if sourceFile == nil || !sourceFile.IsBound() {
		return nil
	}
	return sourceFile
}

// speculationWorkers is how many files speculateFormat formats at once: a quarter of the cores, and none
// when the parallel pass is off. The walk and the type check have every core besides, and the type check is
// what a cold run waits on once formatting has left the critical path. Measured on ahra (#679s763), load 7 to
// 15, interleaved: 4 workers 3.02s median against 3.19s without speculating, 8 workers 3.12s and 16 workers
// 3.13s, the type check slowing from 1.6s to 2.2s as more formatting competed with it.
func speculationWorkers() int {
	if parallelFormatWorkers() == 0 {
		return 0
	}
	return max(runtime.GOMAXPROCS(0)/4, 1)
}

// stop has the workers finish the files they are on and begin no more. Called once the walk is over on a run
// whose walk also did the type check (program.FusedCheck): nothing else then wants the cores, and a quarter of
// them formatting the rest would set the run's end (#679s763).
// narrow has all but the first workers finish the file they are on and take no more. The files left go to the
// workers that remain, and what they have not begun when the walk ends goes to formatInParallel.
func (speculation *formatSpeculation) narrow(workers int) {
	speculation.limit.Store(int32(workers))
}

func (speculation *formatSpeculation) stop() {
	speculation.stopOnce.Do(func() { close(speculation.stopped) })
}

// wait returns once every attempt is in.
func (speculation *formatSpeculation) wait() {
	<-speculation.done
}

// keepable waits for the speculation and returns the attempts that stand: a file no rule proposed a fix for,
// whose bytes, where the walk read the file, are the bytes the attempt read. A result whose file has a
// proposal describes text the fixes will replace, and one read from other bytes than the walk's would pair
// the walk's findings with formatting of a different file.
func (speculation *formatSpeculation) keepable(byFileName map[string][]edit.Proposal, graph *program.Graph) map[string]formatAttempt {
	speculation.wait()
	kept := make(map[string]formatAttempt, len(speculation.attempts))
	for fileName, attempt := range speculation.attempts {
		if len(byFileName[fileName]) != 0 {
			continue
		}
		if attempt.err == nil && graph != nil && graph.Program != nil {
			if sourceFile := graph.Program.GetSourceFile(tspath.RootedFilePath(fileName)); sourceFile != nil && sourceFile.Text() != speculation.read[fileName] {
				continue
			}
		}
		kept[fileName] = attempt
	}
	return kept
}

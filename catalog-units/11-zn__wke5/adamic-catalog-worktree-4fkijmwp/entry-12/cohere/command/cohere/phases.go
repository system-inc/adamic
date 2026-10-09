package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/system-inc/cohere/internal/release/packaging"
)

// phaseName is one step of the pipeline, in the order it runs.
type phaseName string

const (
	phaseFix    phaseName = "fix"
	phaseTypes  phaseName = "types"
	phaseLint   phaseName = "lint"
	phaseUnused phaseName = "unused"
)

// phaseOrder is the pipeline, and the order is the design rather than a convenience.
//
//  1. an edit happens
//  2. fix and format   mutates; bails only if the result cannot be parsed
//  3. type check       bails here, alone and loudly
//  4. lint             only reachable once the code means something
//
// Mutation runs first so everything downstream sees the repaired tree: reporting a finding a fixer
// would have silently repaired wastes the reader's attention. Types gate lint because findings
// against wrong semantics are noise a reader has to re-read after fixing the real problem — one type
// error alone at the top beats one type error buried under a hundred style findings in a file that
// does not compile.
//  5. unused           a report, opt-in, and never part of the gate
//
// Unused runs last and only when asked for. It is the one phase a bare `cohere` does not run, and
// that break from the other three is deliberate: the first three answer "is this correct", which
// nobody should have to opt into, while unused answers "is this still wanted", which is a question
// with no right answer that a build can enforce. An export held for an external consumer is unused
// and correct. So it reports, it never fails a build on its own account, and it stays out of the
// way of the gate people run a hundred times a day.
//
// It runs after types for a hard reason rather than a tidy one. The reference analysis is keyed on
// symbol identity, and a tree that does not type-check has unresolved symbols, which read as
// "nothing references this". Reporting live code as unused because the build was broken is the
// failure mode that gets a report closed and never reopened, so the phase states loudly when it ran
// without a clean type check rather than quietly producing a longer list.
var phaseOrder = []phaseName{phaseFix, phaseTypes, phaseLint, phaseUnused}

// optInPhases are the phases a bare `cohere` does not run.
//
// Kept as a set rather than a comparison against phaseUnused so that a second opt-in phase does not
// have to rediscover every place the distinction matters. Three places consult it today, and a
// fourth phase added without noticing all three would be reported wrongly in one of them.
var optInPhases = map[phaseName]bool{phaseUnused: true}

// phaseOutcome is what happened to one phase.
type phaseOutcome string

const (
	// outcomeRan is the only outcome that means the phase's findings can be trusted as complete.
	outcomeRan phaseOutcome = "ran"

	// outcomeSkipped is a phase the caller turned off with a flag.
	outcomeSkipped phaseOutcome = "skipped"

	// outcomeNotReached is a phase that would have run and never got the chance, because an earlier
	// phase bailed.
	//
	// Distinct from skipped, and the distinction is the whole point of this file. A phase nobody
	// asked for and a phase that was cut off by a failure upstream both produce no findings, and a
	// reader who cannot tell them apart reads "no lint findings" off a run where lint never
	// executed.
	outcomeNotReached phaseOutcome = "not reached"

	// outcomeReused is a phase whose work another phase already did, and whose findings are therefore
	// complete without this phase having walked.
	//
	// Distinct from ran for the same reason notReached is distinct from skipped. The lint phase can
	// reuse the fix phase's walk when nothing was rewritten, and reporting that as `ran in 0s` states
	// two false things at once: that the walk happened here, and that it was free. The walk happened,
	// it cost about a second, and it was billed to the phase that performed it.
	//
	// This is the same accounting defect the phase table already had one level down, where every row
	// was conditional on what ran before it. A reused phase reporting a zero would have reintroduced
	// it in a new place: a number that is arithmetically consistent and describes nothing.
	outcomeReused phaseOutcome = "reused"

	// outcomeChecked is the fix phase under `--no-fix`: it ran everything a writing run runs, wrote
	// nothing, and counted the files that would have changed.
	//
	// Distinct from skipped, which is what `--no-fix` used to record, and the difference is a finding.
	// A skipped fix phase looked at nothing, so an unformatted file passed a clean run unseen; a
	// checked one looked at every candidate, and its count is in the verdict.
	outcomeChecked phaseOutcome = "checked"
)

// phaseRecord is one phase's outcome, and what it cost.
type phaseRecord struct {
	Name     phaseName
	Outcome  phaseOutcome
	Elapsed  time.Duration
	Detail   string
	Findings int

	// Narrowed says which part of the phase ran when a flag narrowed it, printed beside the phase's time
	// so a narrowed run cannot read as the whole phase. See narrow.
	Narrowed string
}

// pipelineReport is what the whole run did, phase by phase.
//
// This exists because the failure this tool was built to eliminate is a confident verdict over work
// that was never done. A run that bailed at types and a run that linted cleanly both print no lint
// findings, and only a statement of which phases actually executed separates them. So the phase line
// prints on every run, including the successful ones, and a bail is never silent.
type pipelineReport struct {
	records []phaseRecord
	// graph is how long the type graph took to build, which is a phase in every sense that matters
	// to a reader deciding where the time went, and is not one of the records because nothing can
	// skip it.
	graph time.Duration
	// filesInScope and filesInProgram say how much of the tree this run visited.
	//
	// Zero means the run was not scoped and the coverage line has nothing to add. Anything else is a
	// deliberate narrowing whose size the phase lines cannot express: they report which phases ran,
	// and a run that checked one file out of 3,542 in every phase looks complete to them.
	filesInScope   int
	filesInProgram int

	// graphSkipped is a run that built no graph on purpose (`--format-only`), so the accounting says so
	// rather than reporting a zero-length build.
	graphSkipped bool

	// processStart is when the process began, as close to it as a Go program can observe.
	//
	// Recorded so the phase line can state its own completeness. Summing the phases and calling it
	// the run time is the same error as summing the rules and calling it coverage: it reports what
	// was measured as though it were everything that happened. On this tree the phases account for
	// about two thirds of the wall clock, and the missing third was invisible until this field
	// existed.
	processStart time.Time

	// cacheOff is `--no-cache`, said on every such run so a cold number cannot pass for a warm one.
	cacheOff bool

	// contentKeyedShapes is how many files' shapes this run keyed on their content, for want of a recorded
	// declaration signature. See program.Graph.Signatures.
	contentKeyedShapes int

	// rootNote says which project was checked when that is not the directory the run started in.
	// Empty otherwise. See projectLocation.rootNote.
	rootNote string

	// graphNotBuilt is set when the run ended before the graph was needed, so the accounting line
	// says so rather than reporting a graph that took 0s.
	graphNotBuilt bool

	// graphLabel and graphNotBuiltSentence name the step nothing can skip, in the accounting line.
	// Empty means a TypeScript run, whose step is building the graph. A Swift run's is describing the
	// package, and its accounting says so rather than billing a graph that was never built.
	graphLabel            string
	graphNotBuiltSentence string

	// engineSourceTreeModified is set when an external engine reports that it, too, was built from a
	// modified tree. Its findings are the run's findings, so its provenance is the run's provenance.
	engineSourceTreeModified bool

	// incompleteBeyondPhases says what was not checked when every phase ran and the run still fell
	// short: a file the compiler left no record for, a file a rule crashed on, a file the engine could
	// not read. Empty when nothing like that happened. The phase line cannot express it, so the
	// coverage warning names it instead of pointing at phases that all say they ran.
	incompleteBeyondPhases string
}

// record notes what a phase did. Called once per phase, in order.
func (r *pipelineReport) record(name phaseName, outcome phaseOutcome, elapsed time.Duration, detail string) {
	r.records = append(r.records, phaseRecord{
		Name:    name,
		Outcome: outcome,
		Elapsed: elapsed,
		Detail:  detail,
	})
}

// recordChecked notes the fix phase under `--no-fix`: it ran, wrote nothing, and found this many files
// that a writing run would change.
func (r *pipelineReport) recordChecked(name phaseName, elapsed time.Duration, wouldChange int) {
	r.records = append(r.records, phaseRecord{
		Name:     name,
		Outcome:  outcomeChecked,
		Elapsed:  elapsed,
		Findings: wouldChange,
	})
}

// narrow notes that only part of an already recorded phase ran, such as the fix phase under
// `--format-only`, which formats and proposes no fixes.
func (r *pipelineReport) narrow(name phaseName, part string) {
	for index := range r.records {
		if r.records[index].Name == name {
			r.records[index].Narrowed = part
		}
	}
}

// requested records which phases the caller actually asked for.
//
// Only consulted for the opt-in phases. A phase that runs by default is always "requested" in the
// sense that matters here: nobody had to ask, so being cut off by a bail is always a gap.
type requestedPhases map[phaseName]bool

// markRemainingNotReached fills in every phase after the one that bailed.
//
// Called at the bail rather than left implicit, because a phase that is simply absent from the
// report is indistinguishable from one the reporter forgot. Every phase in the pipeline appears in
// every run's output, with a stated outcome.
//
// An opt-in phase nobody asked for is recorded as skipped rather than as not reached, and the
// distinction is the same one this whole file exists to preserve. "Did not run because types
// bailed" tells the reader something was taken from them; "not requested" tells them nothing was.
// Printing the first when the second is true manufactures a gap out of an ordinary run, which is
// the mirror image of the silent-green failure and just as misleading.
func (r *pipelineReport) markRemainingNotReached(bailedAt phaseName, reason string) {
	r.markRemainingNotReachedFor(bailedAt, reason, nil)
}

// markRemainingNotReachedFor is markRemainingNotReached told which opt-in phases were asked for.
func (r *pipelineReport) markRemainingNotReachedFor(bailedAt phaseName, reason string, requested requestedPhases) {
	reached := false
	for _, name := range phaseOrder {
		if name == bailedAt {
			reached = true
			continue
		}
		if !reached {
			continue
		}
		if r.has(name) {
			continue
		}
		// An opt-in phase nobody named was not deprived of anything by the bail.
		if optInPhases[name] && !requested[name] {
			r.record(name, outcomeSkipped, 0, "not requested — this is a report, ask for it with --unused")
			continue
		}
		r.record(name, outcomeNotReached, 0, fmt.Sprintf("%s bailed: %s", bailedAt, reason))
	}
}

func (r *pipelineReport) has(name phaseName) bool {
	for _, record := range r.records {
		if record.Name == name {
			return true
		}
	}
	return false
}

// checkedEverything reports whether every phase that reports findings actually ran.
//
// The warning this drives is about coverage of the *checks*, so it asks whether anything that could
// have found a problem was prevented from looking. That is a narrower question than "did all three
// phases run", and the difference is the fix phase: a run that checked types and lint completely is
// not a partial check merely because it was told not to fix (`--lint`, `--types`). A warning that
// fired there would appear on every narrowed run, and a warning that fires when nothing is wrong is
// one people learn to stop reading — which would cost exactly the case it exists for. Under
// `--no-fix` the fix phase is not skipped at all: it is checked, and its count is in the verdict.
//
// A phase that was cut off is always a gap, including the fix phase, because a bail means the run
// stopped early rather than chose not to act.
func (r *pipelineReport) checkedEverything() bool {
	for _, record := range r.records {
		if record.Outcome == outcomeNotReached {
			return false
		}
		// A reporting phase that never ran is a hole in the verdict. Fix is exempt when it was
		// skipped on purpose: declining to mutate withholds no finding.
		//
		// Unused is exempt for a different reason, and it is worth stating rather than folding into
		// the same clause. Fix is exempt because skipping it withholds no finding; unused withholds
		// findings by definition when it is skipped. It is exempt because being absent is its NORMAL
		// state — it is opt-in, so a bare `cohere` skips it every time, and a coverage warning that
		// fires on every ordinary run is one people learn to stop reading. That would cost exactly
		// the case the warning exists for.
		//
		// What keeps this honest is that the phase line still prints `unused skipped (not
		// requested)` on every run, so the absence is stated even though it is not warned about. The
		// warning is for a gap somebody did not choose; this one is chosen by default.
		if record.Outcome == outcomeSkipped && record.Name != phaseFix && !optInPhases[record.Name] {
			return false
		}
	}
	return true
}

// detailNotRequested is the skip detail that means the caller did not ask for the phase, rather than that
// something stopped it. The unused phase's spelling extends it with how to ask, after a dash.
const detailNotRequested = "not requested"

// skippedAsNotRequested reports whether a phase was skipped because nobody asked for it.
func skippedAsNotRequested(record phaseRecord) bool {
	return record.Outcome == outcomeSkipped &&
		(record.Detail == detailNotRequested || strings.HasPrefix(record.Detail, detailNotRequested+" — "))
}

// checkedEveryRequestedPhase reports whether every phase the caller asked for ran, which is what an
// external engine's `complete` claims (swift/Contract.md).
//
// It is checkedEverything with one difference: a phase skipped as not requested is not a gap, since
// `--lint` asked for lint and got all of it. Every other skip of a reporting phase still is one, and so
// is anything not reached. checkedEverything keeps its wider answer for the phase line, which says on a
// narrowed run that the phases above did not check everything, true of `--lint` in either engine.
func (r *pipelineReport) checkedEveryRequestedPhase() bool {
	for _, record := range r.records {
		if record.Outcome == outcomeNotReached {
			return false
		}
		if record.Outcome == outcomeSkipped && record.Name != phaseFix && !optInPhases[record.Name] &&
			!skippedAsNotRequested(record) {
			return false
		}
	}
	return true
}

// narrowedNote is a record's narrowing in parentheses, or nothing for a whole phase.
func narrowedNote(record phaseRecord) string {
	if record.Narrowed == "" {
		return ""
	}
	return " (" + record.Narrowed + ")"
}

// Write prints the phase line: which phases ran, which did not, and why.
//
// It prints on every run rather than only on failures. The whole argument of this project is that a
// run which checked nothing must not be able to print like a run that checked everything and found
// it clean, and a phase summary that only appears when something went wrong reintroduces exactly
// that ambiguity for the successful case.
func (r *pipelineReport) Write(out io.Writer) {
	parts := make([]string, 0, len(r.records))
	for _, name := range phaseOrder {
		for _, record := range r.records {
			if record.Name != name {
				continue
			}
			switch record.Outcome {
			case outcomeRan:
				parts = append(parts, fmt.Sprintf("%s ran in %s%s", record.Name, round(record.Elapsed), narrowedNote(record)))
			case outcomeSkipped:
				parts = append(parts, fmt.Sprintf("%s skipped (%s)", record.Name, record.Detail))
			case outcomeNotReached:
				parts = append(parts, fmt.Sprintf("%s did not run (%s)", record.Name, record.Detail))
			case outcomeReused:
				// Says whose work it was, because "reused" without a source is an unfalsifiable claim:
				// a reader cannot check a number that names no owner.
				parts = append(parts, fmt.Sprintf("%s reused %s", record.Name, record.Detail))
			case outcomeChecked:
				// The count rides on the phase, zero included, so a clean `--no-fix` run states that it
				// looked and found nothing to change rather than leaving that to be inferred.
				parts = append(parts, fmt.Sprintf("%s checked in %s%s, %d file%s would change",
					record.Name, round(record.Elapsed), narrowedNote(record), record.Findings, plural(record.Findings)))
			}
		}
	}

	fmt.Fprintf(invocationOutput(out), "phases: %s\n", strings.Join(parts, " · "))

	r.writeAccounting(out)
	// Under the same condition as the total: both describe the process, and a report with no process
	// behind it has neither.
	if !r.processStart.IsZero() {
		fmt.Fprintf(invocationOutput(out), "  %s\n", activeMemoryPolicy.line())
	}

	// The explicit sentence for the case that matters most. A reader who takes only the last line
	// away from a bailed run must not take away a clean bill of health.
	//
	// It fires on a genuine gap in what was checked, not on every partial run. A skipped fix phase
	// withholds no finding, and `--lint` alone is a deliberate narrowing whose own phase line already says so;
	// warning on those would put the sentence on ordinary runs until people stopped reading it,
	// which would cost exactly the case it exists for.
	// A binary built from a modified tree reproduces no commit, so a number it produced cannot be
	// checked against anything later.
	//
	// Said here rather than only under `--version`, because the run that needs to disclose it is the
	// one producing findings, and nobody types `--version` before reading a count. It caught a real
	// case: after reverting an experiment I never rebuilt, so four rounds of parity came from a
	// binary carrying code that no longer existed. The numbers happened to be unaffected, and the
	// only reason I knew that is that I checked the version line by hand.
	//
	// Silent on a clean build. A line on every ordinary run is one people learn to skip, which would
	// cost exactly the case it exists for.
	r.writeProvenanceWarning(out, release.Current().SourceTreeModified || r.engineSourceTreeModified)

	if r.rootNote != "" {
		fmt.Fprintf(out, "  %s\n", r.rootNote)
	}
	if r.cacheOff {
		fmt.Fprintln(out, "  cache: off, by --no-cache: nothing was read from or written to this project's cache table or its incremental build info")
	}
	// Said, because keying a shape on content is a choice with a cost: an edit to one of these files re-runs
	// its importers' shape-keyed rules. The alternative, a declaration emit up front, has no ceiling (#5txm9gg).
	if r.contentKeyedShapes > 0 {
		subject := "files' shapes are"
		if r.contentKeyedShapes == 1 {
			subject = "file's shape is"
		}
		fmt.Fprintf(out, "  cache: %d %s keyed on content, with no recorded declaration signature (computing one up front has no ceiling), so an edit to one re-runs its importers' shape-keyed rules\n",
			r.contentKeyedShapes, subject)
	}

	// File scope is its own dimension and the phase lines cannot express it. A run narrowed to one
	// file ran every phase it was asked for, so `the phases above say what was not checked` points a
	// reader at lines that answer a different question.
	//
	// Said first, and with both numbers, because it is the larger gap: skipping a phase withholds one
	// kind of finding, and checking 1 file of 3,542 withholds every kind on 3,541 files.
	if r.filesInScope > 0 && r.filesInScope < r.filesInProgram {
		fmt.Fprintf(out, "  this run checked %d of %d files — the rest were not looked at\n",
			r.filesInScope, r.filesInProgram)
	}

	if !r.checkedEverything() {
		fmt.Fprintf(out, "  this run did not check everything — the phases above say what was not checked\n")
	} else if r.incompleteBeyondPhases != "" {
		fmt.Fprintf(out, "  this run did not check everything — %s\n", r.incompleteBeyondPhases)
	}
}

// writeAccounting states how much of the run the phases above actually explain.
//
// A timing line that reports only its own phases invites the reader to add them up and treat the
// total as the run, which is wrong here by roughly a third. The gap is real work — process start,
// configuration, walking the file scope, printing findings, exit — and none of it is attributable
// to a phase, so leaving it unnamed makes it unownable. Naming it is the same discipline as the
// coverage line: report what was not measured, rather than let a partial account read as a full one.
//
// It prints nothing when the process start was never recorded, because a computed total that is
// silently measuring from the zero time would be worse than no line at all.
func (r *pipelineReport) writeAccounting(out io.Writer) {
	if r.processStart.IsZero() {
		return
	}

	total := time.Since(r.processStart)
	graphLabel, graphNotBuiltSentence := "graph", "no graph was built and no phase ran"
	if r.graphLabel != "" {
		graphLabel, graphNotBuiltSentence = r.graphLabel, r.graphNotBuiltSentence
	}
	if r.graphNotBuilt {
		fmt.Fprintf(invocationOutput(out), "  total %s — %s\n", round(total), graphNotBuiltSentence)
		return
	}
	if r.graphSkipped {
		phases := time.Duration(0)
		for _, record := range r.records {
			if record.Outcome == outcomeRan || record.Outcome == outcomeChecked {
				phases += record.Elapsed
			}
		}
		fmt.Fprintf(invocationOutput(out), "  total %s — no graph built (formatting reads none), phases %s, %s outside any phase\n",
			round(total), round(phases), round(max(total-phases, 0)))
		return
	}
	accounted := r.graph
	for _, record := range r.records {
		// Only phases that actually spent the time are summed. A reused phase records a zero elapsed
		// because its work is already counted under the phase that performed it, and adding anything
		// for it here would double-count a single walk across two rows.
		if record.Outcome == outcomeRan || record.Outcome == outcomeChecked {
			accounted += record.Elapsed
		}
	}

	unaccounted := total - accounted
	if unaccounted < 0 {
		// The phases run concurrently with each other in places, so a sum can exceed the wall
		// clock. Reporting a negative remainder would be nonsense; reporting the overlap honestly
		// is the useful reading.
		fmt.Fprintf(invocationOutput(out), "  total %s — %s %s plus phases, overlapping by %s\n",
			round(total), graphLabel, round(r.graph), round(-unaccounted))
		return
	}

	fmt.Fprintf(invocationOutput(out), "  total %s — %s %s, phases %s, %s outside any phase\n",
		round(total), graphLabel, round(r.graph), round(accounted-r.graph), round(unaccounted))
}

// writeProvenanceWarning says when the binary cannot be traced to a commit.
//
// Takes the flag rather than reading it, so a fixture can hold both directions. `Current()` reads
// stamps the linker wrote, which a test cannot vary, and a test that could only observe the build
// it happens to run under would assert nothing.
func (r *pipelineReport) writeProvenanceWarning(out io.Writer, sourceTreeModified bool) {
	if !sourceTreeModified {
		return
	}
	fmt.Fprintf(out, "  this binary was built from a modified tree, so no commit reproduces these findings\n")
}

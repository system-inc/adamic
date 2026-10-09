package main

import (
	"strings"
	"testing"
	"time"
)

// render is the whole phase line, which is what a person actually reads.
//
// Asserted whole rather than by substring, deliberately. A defect in this package's summary line
// survived a suite of substring assertions earlier tonight: every clause was present, every number
// was correct, and the sentence they composed was wrong. A fragment check cannot see that.
func render(report *pipelineReport) string {
	builder := &strings.Builder{}
	report.Write(builder)
	return builder.String()
}

// A run where all three phases ran says so, and says nothing else.
func TestACompleteRunPrintsNoWarning(t *testing.T) {
	t.Parallel()
	report := &pipelineReport{}
	report.record(phaseFix, outcomeRan, 4*time.Millisecond, "")
	report.record(phaseTypes, outcomeRan, 12*time.Millisecond, "")
	report.record(phaseLint, outcomeRan, 7*time.Millisecond, "")

	got := render(report)
	want := "phases: fix ran in 4ms · types ran in 12ms · lint ran in 7ms\n"
	if got != want {
		t.Fatalf("the phase line reads wrong:\n  want %q\n  got  %q", want, got)
	}
}

// A phase cut off by an upstream bail must be reported as not-run, with the reason, and must warn.
//
// This is the case the whole file exists for. A run that bailed at types and a run that linted
// cleanly both produce zero lint findings, and only this line separates them.
func TestABailNamesTheUnreachedPhasesAndWarns(t *testing.T) {
	t.Parallel()
	report := &pipelineReport{}
	report.record(phaseFix, outcomeRan, 4*time.Millisecond, "")
	report.record(phaseTypes, outcomeRan, 12*time.Millisecond, "")
	report.markRemainingNotReached(phaseTypes, "4 type diagnostics")

	got := render(report)
	// Unused reads as skipped rather than as not reached, because nobody asked for it here. A bail
	// takes nothing from an opt-in phase that was never requested, and saying it did would
	// manufacture a gap out of an ordinary run.
	want := "phases: fix ran in 4ms · types ran in 12ms · lint did not run (types bailed: 4 type diagnostics)" +
		" · unused skipped (not requested — this is a report, ask for it with --unused)\n" +
		"  this run did not check everything — the phases above say what was not checked\n"
	if got != want {
		t.Fatalf("the bail line reads wrong:\n  want %q\n  got  %q", want, got)
	}
}

// An opt-in phase the caller DID ask for and that a bail then cut off is a real gap, and must read
// as one.
//
// This is the other half of the distinction above, and it is the half that would be easy to lose:
// exempting unused from the coverage warning is correct when nobody asked for it and wrong when
// somebody did. Somebody who typed --unused and got no report has had something withheld.
func TestARequestedUnusedPhaseCutOffByABailIsAGap(t *testing.T) {
	t.Parallel()
	report := &pipelineReport{}
	report.record(phaseFix, outcomeRan, 4*time.Millisecond, "")
	report.record(phaseTypes, outcomeRan, 12*time.Millisecond, "")
	report.markRemainingNotReachedFor(phaseTypes, "4 type diagnostics", requestedPhases{phaseUnused: true})

	got := render(report)
	if !strings.Contains(got, "unused did not run (types bailed: 4 type diagnostics)") {
		t.Fatalf("a requested unused phase that was cut off did not read as cut off: %q", got)
	}
	if !warnsAboutCoverage(got) {
		t.Fatalf("a requested phase that was withheld did not warn about coverage: %q", got)
	}
}

// A bare run must not warn merely because unused was not requested, or the warning appears on every
// ordinary run and people stop reading it — which costs exactly the case it exists for.
func TestAnUnrequestedUnusedPhaseDoesNotWarn(t *testing.T) {
	t.Parallel()
	report := &pipelineReport{}
	report.record(phaseFix, outcomeRan, time.Millisecond, "")
	report.record(phaseTypes, outcomeRan, time.Millisecond, "")
	report.record(phaseLint, outcomeRan, time.Millisecond, "")
	report.record(phaseUnused, outcomeSkipped, 0, "not requested")

	got := render(report)
	if warnsAboutCoverage(got) {
		t.Fatalf("an ordinary run warned about coverage because unused was not requested: %q", got)
	}
	// The absence is still stated, which is what keeps the exemption honest.
	if !strings.Contains(got, "unused skipped") {
		t.Fatalf("the phase line did not state that unused was skipped: %q", got)
	}
}

// A bail at the first phase must mark everything after it, not just the next one.
func TestABailAtFixMarksBothLaterPhases(t *testing.T) {
	t.Parallel()
	report := &pipelineReport{}
	report.record(phaseFix, outcomeRan, time.Millisecond, "")
	report.markRemainingNotReached(phaseFix, "the rewrite did not parse")

	got := render(report)
	if !strings.Contains(got, "types did not run") || !strings.Contains(got, "lint did not run") {
		t.Fatalf("a bail at fix left a later phase unaccounted for: %q", got)
	}
	if !warnsAboutCoverage(got) {
		t.Fatalf("a bail did not warn: %q", got)
	}
}

// warnsAboutCoverage reports whether the rendered output carries the incomplete-coverage sentence.
func warnsAboutCoverage(line string) bool {
	return strings.Contains(line, "did not check everything")
}

// `--no-fix` checks the fix phase rather than skipping it, and says how many files would change, zero
// included. It must not warn: it withheld nothing. This is the continuous-integration path and a
// warning that fires on every green run is one people stop reading — which costs exactly the case
// the warning exists for.
func TestNoFixChecksAndDoesNotWarn(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		wouldChange int
		want        string
	}{
		{0, "phases: fix checked in 3ms, 0 files would change · types ran in 12ms · lint ran in 7ms\n"},
		{1, "phases: fix checked in 3ms, 1 file would change · types ran in 12ms · lint ran in 7ms\n"},
	} {
		report := &pipelineReport{}
		report.recordChecked(phaseFix, 3*time.Millisecond, testCase.wouldChange)
		report.record(phaseTypes, outcomeRan, 12*time.Millisecond, "")
		report.record(phaseLint, outcomeRan, 7*time.Millisecond, "")

		if got := render(report); got != testCase.want {
			t.Fatalf("--no-fix warned or read wrong:\n  want %q\n  got  %q", testCase.want, got)
		}
	}
}

// A reporting phase that was turned off is a real hole in the verdict and must warn, because its
// findings are absent from the output and a reader cannot tell that from having none.
func TestSkippingAReportingPhaseWarns(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name    string
		skipped phaseName
	}{
		{"types skipped", phaseTypes},
		{"lint skipped", phaseLint},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			report := &pipelineReport{}
			for _, name := range phaseOrder {
				if name == testCase.skipped {
					report.record(name, outcomeSkipped, 0, "not requested")
					continue
				}
				report.record(name, outcomeRan, time.Millisecond, "")
			}

			if !warnsAboutCoverage(render(report)) {
				t.Fatalf("skipping %s did not warn: %q", testCase.skipped, render(report))
			}
		})
	}
}

// The warning must be able to fire and able to stay silent. A check that has only ever done one of
// those has not been shown to discriminate.
func TestTheWarningDiscriminates(t *testing.T) {
	t.Parallel()
	complete := &pipelineReport{}
	for _, name := range phaseOrder {
		complete.record(name, outcomeRan, time.Millisecond, "")
	}

	incomplete := &pipelineReport{}
	incomplete.record(phaseFix, outcomeRan, time.Millisecond, "")
	incomplete.markRemainingNotReached(phaseFix, "control")

	if warnsAboutCoverage(render(complete)) {
		t.Fatalf("the warning fired on a complete run")
	}
	if !warnsAboutCoverage(render(incomplete)) {
		t.Fatalf("the warning did not fire on an incomplete run")
	}
}

// Phases print in pipeline order regardless of the order they were recorded in. The line is read as
// a sequence, so a report that listed them by completion time would misdescribe the pipeline.
func TestPhasesPrintInPipelineOrder(t *testing.T) {
	t.Parallel()
	report := &pipelineReport{}
	report.record(phaseLint, outcomeRan, time.Millisecond, "")
	report.record(phaseFix, outcomeRan, time.Millisecond, "")
	report.record(phaseTypes, outcomeRan, time.Millisecond, "")

	got := render(report)
	fixAt := strings.Index(got, "fix ran")
	typesAt := strings.Index(got, "types ran")
	lintAt := strings.Index(got, "lint ran")
	if !(fixAt < typesAt && typesAt < lintAt) {
		t.Fatalf("phases printed out of pipeline order: %q", got)
	}
}

// markRemainingNotReached must never overwrite a phase that already reported an outcome. The phase
// that bailed ran, and recording it twice would print it twice.
func TestMarkingDoesNotDuplicateAnAlreadyRecordedPhase(t *testing.T) {
	t.Parallel()
	report := &pipelineReport{}
	report.record(phaseFix, outcomeRan, time.Millisecond, "")
	report.record(phaseTypes, outcomeRan, time.Millisecond, "")
	report.record(phaseLint, outcomeRan, time.Millisecond, "")
	report.markRemainingNotReached(phaseFix, "should change nothing")

	got := render(report)
	if strings.Count(got, "types") != 1 || strings.Count(got, "lint") != 1 {
		t.Fatalf("a phase was recorded twice: %q", got)
	}
	if warnsAboutCoverage(got) {
		t.Fatalf("marking after the fact turned a complete run incomplete: %q", got)
	}
}

// A reused phase says whose work it reused, and does not claim to have run in zero time.
//
// `lint ran in 0s` was the first version of this line and it stated two false things at once: that
// the walk happened in the lint phase, and that it was free. The walk happened in the fix phase and
// cost about a second. The counts printed beside it were real, which is what made the zero hard to
// see — everything around it checked out.
func TestAReusedPhaseNamesWhatItReused(t *testing.T) {
	t.Parallel()
	report := &pipelineReport{}
	report.record(phaseFix, outcomeRan, 1500*time.Millisecond, "")
	report.record(phaseTypes, outcomeRan, 200*time.Millisecond, "")
	report.record(phaseLint, outcomeReused, 0, "the fix phase's walk (nothing was rewritten)")

	got := render(report)
	want := "phases: fix ran in 1.5s · types ran in 200ms · lint reused the fix phase's walk (nothing was rewritten)\n"
	if got != want {
		t.Fatalf("the phase line reads wrong:\n  want %q\n  got  %q", want, got)
	}
}

// A reused phase is a complete phase, so it must not warn.
//
// Reused and not-reached both leave a phase having done no walking of its own, and they mean
// opposite things: one has findings that are complete, the other has none at all. A warning on a
// reused phase would send a reader looking for a gap that is not there, which is the same
// false-alarm failure this file's warnings exist to avoid in the other direction.
func TestAReusedPhaseDoesNotWarn(t *testing.T) {
	t.Parallel()
	report := &pipelineReport{}
	report.record(phaseFix, outcomeRan, 1500*time.Millisecond, "")
	report.record(phaseTypes, outcomeRan, 200*time.Millisecond, "")
	report.record(phaseLint, outcomeReused, 0, "the fix phase's walk (nothing was rewritten)")

	if got := render(report); strings.Contains(got, "warning") {
		t.Fatalf("a reused phase warned, but its findings are complete:\n  %q", got)
	}
}

// A reused phase contributes nothing to the accounted total, because its work is already counted.
//
// One walk billed to two rows is the defect this whole change exists to remove. Adding a reused
// phase's elapsed time back into the sum would reintroduce it in the accounting line while the
// phase line reads correctly, which is the harder half to notice.
func TestAReusedPhaseIsNotDoubleCounted(t *testing.T) {
	t.Parallel()
	withReuse := &pipelineReport{processStart: time.Now().Add(-3 * time.Second)}
	withReuse.graph = 500 * time.Millisecond
	withReuse.record(phaseFix, outcomeRan, 1500*time.Millisecond, "")
	withReuse.record(phaseLint, outcomeReused, 1500*time.Millisecond, "the fix phase's walk")

	// The reused row carries a non-zero elapsed on purpose: even if a caller records one, the
	// accounting must ignore it rather than trust it.
	//
	// Asserted on the exact phases figure rather than on the absence of a wrong one. The first
	// version of this check looked for a string the buggy version would not have printed either, so
	// it passed against the defect it was written to catch — a guard that could not fail, which is
	// the shape this tree keeps producing. The value is checked directly now: fix's 1.5s and nothing
	// from the reused row.
	got := render(withReuse)
	if !strings.Contains(got, "phases 1.5s") {
		t.Fatalf("expected the accounted phases to be fix's 1.5s alone, with the reused row "+
			"contributing nothing:\n  %q", got)
	}
}

// TestCoverageNamesTheFilesItDidNotLookAt holds the dimension the phase lines cannot express.
//
// A run narrowed to one file runs every phase it was asked for, so the phase lines report a
// complete run and the existing sentence points at them. On the ahra tree that reads as
// `this run did not check everything — the phases above say what was not checked` for a run that
// visited 1 file of 3,542, and the phases say nothing about the other 3,541.
//
// The gap is the larger of the two: skipping a phase withholds one kind of finding, and skipping
// files withholds every kind on every file skipped.
func TestCoverageNamesTheFilesItDidNotLookAt(t *testing.T) {
	t.Parallel()
	scoped := &pipelineReport{filesInScope: 5, filesInProgram: 3542}
	scoped.record(phaseLint, outcomeRan, 8*time.Millisecond, "")

	if rendered := render(scoped); !strings.Contains(rendered, "checked 5 of 3542 files") {
		t.Errorf("a scoped run should name both numbers:\n%s", rendered)
	}

	// The control, and it is what stops this from becoming a line on every run. A run over the whole
	// program has nothing to disclose, and a warning that fires always is one people stop reading.
	whole := &pipelineReport{filesInScope: 3542, filesInProgram: 3542}
	whole.record(phaseLint, outcomeRan, 8*time.Millisecond, "")

	if rendered := render(whole); strings.Contains(rendered, "of 3542 files") {
		t.Errorf("an unscoped run should say nothing about file scope:\n%s", rendered)
	}

	// A report that was never told the numbers says nothing either, rather than reporting 0 of 0.
	silent := &pipelineReport{}
	silent.record(phaseLint, outcomeRan, 8*time.Millisecond, "")

	if rendered := render(silent); strings.Contains(rendered, "files — the rest") {
		t.Errorf("a report with no scope numbers should not invent them:\n%s", rendered)
	}
}

// TestProvenanceWarningFiresOnlyOnAModifiedTree holds the disclosure that a binary cannot be traced
// to a commit.
//
// It caught a real case before it existed. After reverting an experiment I never rebuilt, so four
// rounds of parity measurement came from a binary carrying code that no longer existed in the tree.
// The numbers happened to be unaffected, and the only reason that was knowable is that the version
// line was read by hand.
//
// Both directions, because a warning on every run is one people learn to skip, which would cost
// exactly the case it exists for.
func TestProvenanceWarningFiresOnlyOnAModifiedTree(t *testing.T) {
	t.Parallel()
	report := &pipelineReport{}

	modified := &strings.Builder{}
	report.writeProvenanceWarning(modified, true)
	if !strings.Contains(modified.String(), "no commit reproduces these findings") {
		t.Errorf("a modified tree should disclose itself: %q", modified.String())
	}

	clean := &strings.Builder{}
	report.writeProvenanceWarning(clean, false)
	if clean.String() != "" {
		t.Errorf("a clean build should say nothing: %q", clean.String())
	}
}

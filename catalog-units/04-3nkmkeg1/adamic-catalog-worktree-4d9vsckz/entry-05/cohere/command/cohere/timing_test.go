package main

import (
	"strings"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/types/program"
)

// TestPairsWithEqualNodesAndUnequalCostAreFlagged covers the shape that found the attribution bug.
//
// Two rules offered exactly the same node count and differing by orders of magnitude means one is
// carrying work the other reads for free. Requiring three or more rules would have missed the case
// that motivated this note: it was found on a pair, 3,407 nodes each, 83x apart.
func TestPairsWithEqualNodesAndUnequalCostAreFlagged(t *testing.T) {
	t.Parallel()
	timings := program.NewTimings([]string{"expensive", "cheap"})
	setCost(timings, "expensive", 100*time.Millisecond, 3407)
	setCost(timings, "cheap", time.Millisecond, 3407)

	rendered := renderTimings(timings, 200*time.Millisecond)

	if !strings.Contains(rendered, "differ") {
		t.Fatalf("a pair with equal nodes and 100x cost was not flagged:\n%s", rendered)
	}
	if !strings.Contains(rendered, "3407") {
		t.Fatalf("the note did not name the shared node count:\n%s", rendered)
	}
}

// TestSimilarCostsAreNotFlagged is the half that keeps the note meaningful.
//
// Rules registered for the same kind and costing about the same are unremarkable. A note that fired
// on every pair sharing a node count would be noise, and noise in a diagnostic is how a real signal
// stops being read.
func TestSimilarCostsAreNotFlagged(t *testing.T) {
	t.Parallel()
	timings := program.NewTimings([]string{"first", "second"})
	setCost(timings, "first", 10*time.Millisecond, 3407)
	setCost(timings, "second", 8*time.Millisecond, 3407)

	if rendered := renderTimings(timings, 100*time.Millisecond); strings.Contains(rendered, "differ") {
		t.Fatalf("two comparably-priced rules were flagged as sharing work:\n%s", rendered)
	}
}

// TestDifferentNodeCountsAreNotCompared keeps the note from pairing rules that never registered for
// the same kinds, where a cost difference means nothing.
func TestDifferentNodeCountsAreNotCompared(t *testing.T) {
	t.Parallel()
	timings := program.NewTimings([]string{"many-nodes", "few-nodes"})
	setCost(timings, "many-nodes", 100*time.Millisecond, 600000)
	setCost(timings, "few-nodes", time.Millisecond, 3407)

	if rendered := renderTimings(timings, 200*time.Millisecond); strings.Contains(rendered, "differ") {
		t.Fatalf("rules with unrelated node counts were compared:\n%s", rendered)
	}
}

// TestSharedFillIsReported proves the cost lands on the derivation rather than vanishing.
func TestSharedFillIsReported(t *testing.T) {
	t.Parallel()
	timings := program.NewTimings([]string{"a-rule"})
	setCost(timings, "a-rule", 10*time.Millisecond, 100)
	timings.SetSharedForTest("nexus.allComments", 120*time.Millisecond)

	rendered := renderTimings(timings, 200*time.Millisecond)
	if !strings.Contains(rendered, "shared: nexus.allComments cost 120ms of CPU") {
		t.Fatalf("the shared derivation was not named with its cost:\n%s", rendered)
	}
	if !strings.Contains(rendered, "computed once per file") {
		t.Fatalf("the note did not explain what a shared cost is:\n%s", rendered)
	}
}

func setCost(timings *program.Timings, name string, cost time.Duration, nodes int) {
	timings.SetForTest(name, cost, nodes)
}

func renderTimings(timings *program.Timings, wall time.Duration) string {
	builder := &strings.Builder{}
	printTimings(builder, timings, wall)
	return builder.String()
}

// TestTheNamedRulesAreTheRulesWhoseCostsAreQuoted is the assertion the first version of this file
// was missing, and the omission let a real bug through.
//
// The original note printed names[0] and names[len-1], which are list positions, while the costs
// came from a scan for the minimum and maximum. With two rules those coincide, so every fixture
// passed. With three they come apart: the line named a 50ms rule, quoted 400ms, and the rule that
// actually cost 400ms did not appear at all.
//
// A fixture asserting that the note fired is not a fixture asserting the note is right. The earlier
// tests checked for the substring "differ" and for one rule name, both present in a line that was
// wrong about everything else. This one requires the quoted numbers to belong to the named rules.
func TestTheNamedRulesAreTheRulesWhoseCostsAreQuoted(t *testing.T) {
	t.Parallel()
	timings := program.NewTimings([]string{"aaa-middling", "mmm-dearest", "zzz-cheapest"})
	setCost(timings, "aaa-middling", 50*time.Millisecond, 3407)
	setCost(timings, "mmm-dearest", 400*time.Millisecond, 3407)
	setCost(timings, "zzz-cheapest", time.Millisecond, 3407)

	rendered := renderTimings(timings, time.Second)

	line := ""
	for _, candidate := range strings.Split(rendered, "\n") {
		if strings.Contains(candidate, "differ") {
			line = candidate
			break
		}
	}
	if line == "" {
		t.Fatalf("three rules at one node count and 400x apart produced no note:\n%s", rendered)
	}

	// The dearest rule must be named, because it is the one a reader should go look at.
	if !strings.Contains(line, "mmm-dearest") {
		t.Fatalf("the note quotes the dearest cost but does not name the dearest rule:\n  %s", line)
	}
	if !strings.Contains(line, "zzz-cheapest") {
		t.Fatalf("the note does not name the cheapest rule:\n  %s", line)
	}
	// And the rule that is neither must not appear, or the line points somewhere misleading.
	if strings.Contains(line, "aaa-middling") {
		t.Fatalf("the note names a middling rule as though it were an extreme:\n  %s", line)
	}
	if !strings.Contains(line, "400ms") || !strings.Contains(line, "1.0ms") {
		t.Fatalf("the note does not quote both extremes:\n  %s", line)
	}
}

// TestTheHeaderNamesItsUnit pins the line that keeps the table from being read as wall time (#8qyzmxw).
//
// The table this replaced printed wall time under the same columns, and the better-tailwindcss family
// read about 1.2s on it where its real cost was about 0.05s. The unit goes in the first line, and the
// measurement's own cost per call in the second, so a reader cannot mistake what a row is.
func TestTheHeaderNamesItsUnit(t *testing.T) {
	t.Parallel()
	timings := program.NewTimings([]string{"a-rule"})
	setCost(timings, "a-rule", 120*time.Millisecond, 100)
	timings.Account.Calls = 1000
	timings.Account.InstrumentCPU = 150 * time.Microsecond

	rendered := renderTimings(timings, time.Second)
	header := strings.SplitN(strings.TrimLeft(rendered, "\n"), "\n", 2)[0]
	if !strings.Contains(header, "rule CPU") || !strings.Contains(header, "not wall time") {
		t.Fatalf("the header does not say the table is CPU and not wall time:\n  %s", header)
	}
	if !strings.Contains(rendered, "1000 measured calls costs about 150ns to measure") {
		t.Fatalf("the header does not state what measuring a call cost:\n%s", rendered)
	}
}

// TestNoClockIsSaidRatherThanPrintedAsZero covers a platform with no thread CPU clock. A table of zeros
// would read as every rule being free.
func TestNoClockIsSaidRatherThanPrintedAsZero(t *testing.T) {
	t.Parallel()
	timings := program.NewTimings([]string{"a-rule"})
	setCost(timings, "a-rule", 0, 100)
	timings.Account.Unavailable = "this platform has no thread CPU clock to read"

	rendered := renderTimings(timings, time.Second)
	if !strings.Contains(rendered, "no rule CPU this run, because this platform has no thread CPU clock to read") {
		t.Fatalf("the table did not say why it has no CPU:\n%s", rendered)
	}
	if strings.Contains(rendered, "0.00ms") || strings.Contains(rendered, "coverage:") {
		t.Fatalf("a table with no clock printed CPU figures anyway:\n%s", rendered)
	}
	if !strings.Contains(rendered, "a-rule") || !strings.Contains(rendered, " 100 ") {
		t.Fatalf("the counts, which were still counted, were dropped:\n%s", rendered)
	}
}

// TestCoverageStatesEveryPartOfTheWalksCPU guards the line that keeps the table honest.
//
// The rows are rule listeners and per-file setup. A reader who sums the share column gets 100 percent and
// concludes the phase is explained, when the walk that offers nodes to those listeners costs CPU too. The
// wall-clock table could only guess at that from wall clock, which absorbed every wait, and in a parallel
// run it vanished entirely. The threads' clocks account for it, so the line states each part.
func TestCoverageStatesEveryPartOfTheWalksCPU(t *testing.T) {
	t.Parallel()
	timings := program.NewTimings([]string{"a-rule"})
	setCost(timings, "a-rule", 300*time.Millisecond, 100)
	timings.SetSharedForTest("hir.Function", 100*time.Millisecond)
	timings.Account.InstrumentCPU = 100 * time.Millisecond
	timings.Account.WalkCPU = 500 * time.Millisecond
	timings.Account.OtherCPU = 250 * time.Millisecond

	var out strings.Builder
	printCoverage(&out, timings)
	rendered := out.String()

	for _, want := range []string{"1000ms", "rules 30%", "shared derivations 10%", "measuring 10%", "50% the walk itself", "250ms of process CPU"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("coverage did not state %q: %q", want, rendered)
		}
	}
}

// TestCoverageSaysNothingRatherThanComputingFromZero covers the degenerate input. Nothing sampled on the
// walk workers means no share to state, and a share computed from a zero would be harder to notice than a
// missing line.
func TestCoverageSaysNothingRatherThanComputingFromZero(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	printCoverage(&out, program.NewTimings([]string{"a-rule"}))
	if out.String() != "" {
		t.Fatalf("coverage printed a claim computed from zero CPU: %q", out.String())
	}
}

// TestTheTableItselfCarriesTheCoverageLine pins the wiring, not just the function.
//
// The tests above call printCoverage directly, which verifies a fragment and reads as though it verifies
// the behavior. It does not: deleting the call from printTimings leaves them green while the table goes
// back to implying it explains the whole phase. Guard present, fixtures passing, behavior gone.
func TestTheTableItselfCarriesTheCoverageLine(t *testing.T) {
	t.Parallel()
	timings := program.NewTimings([]string{"a-rule"})
	setCost(timings, "a-rule", 285*time.Millisecond, 100)
	timings.Account.WalkCPU = 700 * time.Millisecond

	rendered := renderTimings(timings, 1024*time.Millisecond)
	if !strings.Contains(rendered, "coverage:") || !strings.Contains(rendered, "the walk itself") {
		t.Fatalf("the table did not carry a coverage line naming the walk's own CPU:\n%s", rendered)
	}
}

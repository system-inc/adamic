package edit

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// proposal is a fixture shorthand: a rule name and the span it wants to replace.
func proposal(ruleName string, start int, end int, text string) Proposal {
	return Proposal{
		RuleName: ruleName,
		Fix:      rule.Fix{Range: core.NewTextRange(start, end), Text: text},
	}
}

// appliedRules names the rules whose fixes survived, for comparing against an expectation.
func appliedRules(plan Plan) []string {
	names := make([]string, 0, len(plan.Applied))
	for _, applied := range plan.Applied {
		names = append(names, applied.RuleName)
	}
	return names
}

func expectSame(t *testing.T, label string, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: expected %v, got %v", label, want, got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("%s: expected %v, got %v", label, want, got)
		}
	}
}

// Two fixes overlapping the same range must not both land. The earlier one wins by the stated
// ordering, and the loser is reported with the winner named — a refusal nobody can attribute is
// not actionable.
func TestOverlappingFixesDoNotBothApply(t *testing.T) {
	t.Parallel()
	plan := Resolve([]Proposal{
		proposal("rule-a", 10, 20, "AAA"),
		proposal("rule-b", 15, 25, "BBB"),
	})

	expectSame(t, "applied", appliedRules(plan), []string{"rule-a"})

	if len(plan.Rejected) != 1 {
		t.Fatalf("expected 1 rejection, got %d", len(plan.Rejected))
	}
	if plan.Rejected[0].Reason != ReasonOverlap {
		t.Fatalf("expected reason %q, got %q", ReasonOverlap, plan.Rejected[0].Reason)
	}
	if plan.Rejected[0].ConflictsWith != "rule-a" {
		t.Fatalf("expected the rejection to name rule-a as the winner, got %q", plan.Rejected[0].ConflictsWith)
	}
}

// The overlap decision must not depend on the order proposals arrive in. A fixer whose output
// changes with map iteration order produces a file that differs run to run, which is worse than
// either outcome it is choosing between.
func TestOverlapResolutionIsOrderIndependent(t *testing.T) {
	t.Parallel()
	forward := Resolve([]Proposal{
		proposal("rule-a", 10, 20, "AAA"),
		proposal("rule-b", 15, 25, "BBB"),
	})
	backward := Resolve([]Proposal{
		proposal("rule-b", 15, 25, "BBB"),
		proposal("rule-a", 10, 20, "AAA"),
	})

	expectSame(t, "forward", appliedRules(forward), []string{"rule-a"})
	expectSame(t, "backward", appliedRules(backward), []string{"rule-a"})
}

// A fix fully containing another is still an overlap. This is the case a naive "do the ranges share
// a byte" check gets right and a naive "do they start in the same place" check gets wrong.
func TestNestedFixesDoNotBothApply(t *testing.T) {
	t.Parallel()
	plan := Resolve([]Proposal{
		proposal("outer", 10, 40, "OUTER"),
		proposal("inner", 20, 25, "INNER"),
	})

	expectSame(t, "applied", appliedRules(plan), []string{"outer"})
	if len(plan.Rejected) != 1 || plan.Rejected[0].Proposal.RuleName != "inner" {
		t.Fatalf("expected inner to be refused, got %+v", plan.Rejected)
	}
}

// Adjacent fixes are not overlapping and both must land, or the engine would refuse most of the
// real work: two rules fixing consecutive tokens is the common case, not the exception.
func TestAdjacentFixesBothApply(t *testing.T) {
	t.Parallel()
	plan := Resolve([]Proposal{
		proposal("rule-a", 10, 20, "AAA"),
		proposal("rule-b", 20, 30, "BBB"),
	})

	expectSame(t, "applied", appliedRules(plan), []string{"rule-a", "rule-b"})
	if len(plan.Rejected) != 0 {
		t.Fatalf("expected no rejections, got %+v", plan.Rejected)
	}
}

// Two insertions at the same point are an overlap even though neither consumes a byte. Which one
// goes first is unresolvable and the interleaving is invisible in a diff, so the second is refused.
func TestTwoInsertionsAtOnePointDoNotBothApply(t *testing.T) {
	t.Parallel()
	plan := Resolve([]Proposal{
		proposal("rule-a", 10, 10, "AAA"),
		proposal("rule-b", 10, 10, "BBB"),
	})

	if len(plan.Applied) != 1 {
		t.Fatalf("expected 1 applied, got %d: %v", len(plan.Applied), appliedRules(plan))
	}
	if len(plan.Rejected) != 1 || plan.Rejected[0].Reason != ReasonOverlap {
		t.Fatalf("expected 1 overlap rejection, got %+v", plan.Rejected)
	}
}

// Trimming a range past its leading trivia must never turn a genuine overlap into a silent
// non-overlap.
//
// Raised by @system_cohere_lint when it landed the trivia fix: trimmed ranges are strictly
// narrower, so two fixes that used to collide through shared trivia might now be admitted as
// adjacent, and two rules would both rewrite a region the overlap rule previously protected.
//
// Measured rather than assumed, and the answer is that sibling nodes were never overlapping to
// begin with. A node's untrimmed Loc.Pos() is exactly the previous node's End — the trivia between
// two nodes belongs entirely to the second one's leading span, and is never shared. So for
// `alpha + beta`, alpha is [13,19) and beta is [21,26) untrimmed, [14,19) and [22,26) trimmed:
// adjacent in both forms, overlapping in neither. Narrowing moves a start forward into a region no
// other node's range covered, so it cannot create a new gap between two ranges that genuinely
// intersected.
//
// What that leaves is the case worth a fixture: ranges that overlap in their token text overlap
// whether or not trivia is trimmed, because trimming only ever moves a start rightward and never
// moves an end. The assertion below is on that invariant rather than on any particular parse.
func TestTrimmingAStartCannotDissolveARealOverlap(t *testing.T) {
	t.Parallel()
	// Two fixes whose token spans genuinely intersect.
	untrimmed := Resolve([]Proposal{
		proposal("rule-a", 10, 25, "AAA"),
		proposal("rule-b", 20, 30, "BBB"),
	})
	// The same pair after trimming moved each start forward past its leading trivia. The overlap
	// survives because trimming never moves an end leftward.
	trimmed := Resolve([]Proposal{
		proposal("rule-a", 12, 25, "AAA"),
		proposal("rule-b", 22, 30, "BBB"),
	})

	if len(untrimmed.Rejected) != 1 {
		t.Fatalf("expected the untrimmed pair to collide, got %+v", untrimmed.Rejected)
	}
	if len(trimmed.Rejected) != 1 {
		t.Fatalf("trimming dissolved a real overlap: %+v", trimmed)
	}
	if trimmed.Rejected[0].Reason != ReasonOverlap {
		t.Fatalf("expected an overlap refusal, got %q", trimmed.Rejected[0].Reason)
	}
}

// Three fixes to one file must all land at the right bytes. This is the offset test: applying front
// to back would place the second and third edits wrong by the length delta of everything before
// them, and the failure is silent — the file still parses, it just says something else.
func TestThreeFixesKeepTheirOffsets(t *testing.T) {
	t.Parallel()
	source := "const alpha = 1; const beta = 2; const gamma = 3;"
	//         0123456789...
	//         alpha at 6..11, beta at 23..27, gamma at 39..44

	if source[6:11] != "alpha" || source[23:27] != "beta" || source[39:44] != "gamma" {
		t.Fatalf("fixture offsets are wrong: %q %q %q", source[6:11], source[23:27], source[39:44])
	}

	plan := Resolve([]Proposal{
		// Replacements of different lengths, so a front-to-back application would drift.
		proposal("rule-a", 6, 11, "a"),
		proposal("rule-b", 23, 27, "beeeeta"),
		proposal("rule-c", 39, 44, "g"),
	})

	rewritten, validated := applyToText(source, plan)

	if len(validated.Applied) != 3 {
		t.Fatalf("expected all 3 to apply, got %d (%+v)", len(validated.Applied), validated.Rejected)
	}

	want := "const a = 1; const beeeeta = 2; const g = 3;"
	if rewritten != want {
		t.Fatalf("offsets drifted:\n  want %q\n  got  %q", want, rewritten)
	}
}

// A range that points outside the text is refused rather than applied. A rule computing against a
// stale parse produces exactly this, and slicing on it would panic or silently truncate the file.
func TestOutOfBoundsRangeIsRefused(t *testing.T) {
	t.Parallel()
	source := "const a = 1;"

	rewritten, validated := applyToText(source, Plan{Applied: []Proposal{
		proposal("rule-a", 5, 500, "x"),
	}})

	if len(validated.Applied) != 0 {
		t.Fatalf("expected the out-of-bounds fix to be refused, it applied")
	}
	if len(validated.Rejected) != 1 || validated.Rejected[0].Reason != ReasonInvalidRange {
		t.Fatalf("expected an invalid-range rejection, got %+v", validated.Rejected)
	}
	if rewritten != source {
		t.Fatalf("the text changed despite the refusal: %q", rewritten)
	}
}

// An inverted range is refused for the same reason, and would slice-panic if it were not.
func TestInvertedRangeIsRefused(t *testing.T) {
	t.Parallel()
	source := "const a = 1;"

	_, validated := applyToText(source, Plan{Applied: []Proposal{
		proposal("rule-a", 8, 3, "x"),
	}})

	if len(validated.Rejected) != 1 || validated.Rejected[0].Reason != ReasonInvalidRange {
		t.Fatalf("expected an invalid-range rejection, got %+v", validated.Rejected)
	}
}

// A fix that replaces text with itself is not progress. Counting it as applied would make a run
// that changed nothing report as a run that changed something, and would keep the convergence loop
// spinning on a rule that proposes it every pass.
func TestNoOpFixIsRefused(t *testing.T) {
	t.Parallel()
	source := "const a = 1;"

	rewritten, validated := applyToText(source, Plan{Applied: []Proposal{
		proposal("rule-a", 6, 7, "a"),
	}})

	if len(validated.Applied) != 0 {
		t.Fatalf("expected the no-op fix to be refused, it applied")
	}
	if validated.Rejected[0].Reason != ReasonNoProgress {
		t.Fatalf("expected %q, got %q", ReasonNoProgress, validated.Rejected[0].Reason)
	}
	if rewritten != source {
		t.Fatalf("the text changed on a no-op: %q", rewritten)
	}
}

// Suggestions must never be collected as fixes. The split is intent: a suggestion changes what the
// code means and needs a human to agree, so a fixer that applied them unattended would be changing
// behavior nobody chose.
func TestProposalsFromIgnoresSuggestions(t *testing.T) {
	t.Parallel()
	diagnostics := []rule.Diagnostic{{
		RuleName: "rule-a",
		Fixes:    []rule.Fix{{Range: core.NewTextRange(0, 1), Text: "x"}},
		Suggestions: []rule.Suggestion{{
			Message: rule.Message{Id: "suggest"},
			Fixes:   []rule.Fix{{Range: core.NewTextRange(2, 3), Text: "y"}},
		}},
	}}

	proposals := ProposalsFrom(diagnostics)

	if len(proposals) != 1 {
		t.Fatalf("expected 1 proposal (the fix, not the suggestion), got %d", len(proposals))
	}
	if proposals[0].Fix.Text != "x" {
		t.Fatalf("collected the suggestion instead of the fix: %+v", proposals[0])
	}
}

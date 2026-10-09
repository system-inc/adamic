package react_conformance

import (
	"fmt"
	"sort"
	"strings"
)

// This file is the aggregation layer: it turns "did this rule reproduce this golden" into a number
// whose categories a reader can act on.
//
// # Why the outcome set in score.go was not enough
//
// `Run` distinguishes Passed, Failed, Declined, Excluded, Refused, which is the right partition for
// an implementation that is either present or absent. It is the wrong partition for the question
// this file answers, because it collapses three facts that have different responses:
//
//   - A rule produced nothing because the fixture never reaches its validator upstream either. The
//     response is nothing; the corpus is not asking the question.
//   - A rule produced nothing because the fixture does not import React, so a type-aware rule reads
//     `any` and is correctly silent. The response is to fix the corpus (a stub) or to accept the
//     gap, but it is not to change the rule.
//   - A rule produced nothing because it deliberately does not carry that diagnostic kind, with the
//     boundary written down and machine-checked. The response is to decide whether to widen scope.
//
// All three read as "failed" under a raw score, and the third is the one that matters: it punishes
// exactly the rules that were most careful about stating their limits. A number that does that is
// worse than no number, because it creates pressure to stop writing the limits down.
//
// # The categories, and what each one commits to
//
// Every category here is a claim that can be wrong, and each is testable. `ProveEachCategoryIsReachable`
// exists because a category that can never be non-empty is not a category, it is a comment.

// Verdict is why one fixture landed where it did, for one rule.
type Verdict string

const (
	// VerdictPassed means the rule produced the expected diagnostics.
	VerdictPassed Verdict = "passed"

	// VerdictFailed means the rule ran, the fixture was answerable, and it disagreed with the
	// golden. This is the only category that is a defect in cohere.
	VerdictFailed Verdict = "failed"

	// VerdictUnreachableFixture means the fixture never reaches this validator upstream either.
	//
	// Measured, not assumed: the fixture's expected diagnostics belong to a DIFFERENT rule than the
	// one being scored. React's pipeline aborts earlier, so the golden records the earlier
	// diagnostic and the rule under test is never consulted. The three fixtures named for
	// error-boundaries are the clearest case — all three expect a `Todo` from BuildHIR, because
	// lowering aborts on a `try` with no `catch`.
	VerdictUnreachableFixture Verdict = "declined: unreachable fixture"

	// VerdictUnresolvableTypes means the fixture cannot resolve `@types/react`.
	//
	// A checker-based rule reads `any` for every hook and is correctly silent. Measured over this
	// corpus: 128 of the 202 hook-using fixtures (63%) never import React. A rule can be right on
	// real code and score zero here, and that is a fact about the corpus rather than the rule.
	VerdictUnresolvableTypes Verdict = "declined: unresolvable types"

	// VerdictStatedDivergence means cohere deliberately differs, with the boundary written down.
	//
	// Held separate from failure on purpose. `exhaustive-deps` carries 15 of 17 diagnostic kinds
	// with `TestExhaustiveDepsScopeIsStated` naming the eight dropped cases; `set-state-in-effect`
	// pins a ref exemption as a fixture that fails when it is fixed. Those are decisions, and a
	// score that called them failures would be measuring care as if it were error.
	VerdictStatedDivergence Verdict = "excluded: stated divergence"

	// VerdictNoRuleShipped means no rule in cohere claims this fixture's diagnostics.
	//
	// Separate from unreachable, because the response is different: unreachable means nobody could
	// score it, this means we have not written the rule yet. Six of upstream's 26 rules are the
	// bulk of the corpus and cohere ships none of them.
	VerdictNoRuleShipped Verdict = "declined: no rule shipped"

	// VerdictFlowSyntax means the fixture needs a Flow parser.
	VerdictFlowSyntax Verdict = "excluded: flow syntax"

	// VerdictNotScored means a rule cohere ships owns this fixture but no rule engine was run.
	//
	// This is the honest name for the state the harness is in today, and it is deliberately NOT
	// folded into any decline. A decline says something about the fixture; this says something
	// about the run. Collapsing it into `unreachable` would claim these fixtures cannot be scored,
	// when in fact they are precisely the ones that can — and the addressable set going unmeasured
	// is the single thing a reader most needs to see.
	VerdictNotScored Verdict = "not scored: no rule engine run"
)

// AllVerdicts is every category, in report order.
//
// Ordered rather than ranged over a map, because a scoreboard whose rows move between runs invites
// reading the shape instead of the numbers.
var AllVerdicts = []Verdict{
	VerdictPassed,
	VerdictFailed,
	VerdictUnreachableFixture,
	VerdictUnresolvableTypes,
	VerdictStatedDivergence,
	VerdictNoRuleShipped,
	VerdictFlowSyntax,
	VerdictNotScored,
}

// FixtureVerdict is one fixture's outcome with the evidence for it.
type FixtureVerdict struct {
	Name    string
	Verdict Verdict

	// Reason is why, in one line. Required for every non-pass: a category with no reason is a
	// bucket someone can hide a failure in.
	Reason string

	// Rules are the upstream rule names this fixture's diagnostics belong to.
	Rules []string

	Expected []ExpectedError
	Reported []ReportedError
}

// Report is the whole aggregated run.
type Report struct {
	Considered int
	Counts     map[Verdict]int
	Fixtures   []FixtureVerdict
}

// Check verifies the categories partition the corpus.
//
// Same guard as `Score.Check`, and for the same reason: the arithmetic is the only thing that
// notices a fixture counted twice or not at all.
func (r Report) Check() error {
	sum := 0
	for _, verdict := range AllVerdicts {
		sum += r.Counts[verdict]
	}
	if sum != r.Considered {
		return fmt.Errorf("report does not partition the corpus: considered %d but categories sum to %d", r.Considered, sum)
	}
	if r.Considered != len(r.Fixtures) {
		return fmt.Errorf("considered %d fixtures but recorded %d verdicts", r.Considered, len(r.Fixtures))
	}
	for _, fixtureVerdict := range r.Fixtures {
		if fixtureVerdict.Verdict != VerdictPassed && fixtureVerdict.Reason == "" {
			return fmt.Errorf("%s: %s with no reason; an unreasoned category is somewhere to hide a failure", fixtureVerdict.Name, fixtureVerdict.Verdict)
		}
	}
	return nil
}

// Summary is the scoreboard, one line per category.
//
// It leads with the number and its denominator and then names every category, because the headline
// alone is the thing this whole file exists to stop anyone quoting.
func (r Report) Summary() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "%d / %d passed against facebook/react@%s\n", r.Counts[VerdictPassed], r.Considered, UpstreamSha[:12])
	for _, verdict := range AllVerdicts {
		fmt.Fprintf(&builder, "  %-30s %4d\n", verdict, r.Counts[verdict])
	}
	return builder.String()
}

// ByRule breaks the report down by the upstream rule each fixture belongs to.
func (r Report) ByRule() map[string]map[Verdict]int {
	breakdown := map[string]map[Verdict]int{}
	for _, fixtureVerdict := range r.Fixtures {
		for _, ruleName := range fixtureVerdict.Rules {
			if breakdown[ruleName] == nil {
				breakdown[ruleName] = map[Verdict]int{}
			}
			breakdown[ruleName][fixtureVerdict.Verdict]++
		}
	}
	return breakdown
}

// RuleNames returns the rules in the breakdown, sorted.
func (r Report) RuleNames() []string {
	names := map[string]bool{}
	for _, fixtureVerdict := range r.Fixtures {
		for _, ruleName := range fixtureVerdict.Rules {
			names[ruleName] = true
		}
	}
	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)
	return sorted
}

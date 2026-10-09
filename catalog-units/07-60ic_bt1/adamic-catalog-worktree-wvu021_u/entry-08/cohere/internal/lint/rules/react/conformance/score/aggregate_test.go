package react_conformance_score

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rules/react/conformance"
)

// engineImplementation runs whichever wired rule owns a fixture, and declines the rest.
//
// This is the `react_conformance.Implementation` the harness was written against and never had. A
// fixture whose attributed rule has no wired entry returns ErrUnsupported rather than an empty
// Result, which is the whole distinction `score.go` draws: an empty Result claims "I looked and
// found nothing", and on a corpus where every fixture expects at least one error that claim is a
// wrong answer rather than an absent one.
type engineImplementation struct{ t *testing.T }

func (e engineImplementation) Analyze(fixture react_conformance.Fixture) (react_conformance.Result, error) {
	rules, complete := fixture.Rules()
	if !complete {
		return react_conformance.Result{}, &react_conformance.ErrUnsupported{Reason: "not every diagnostic could be attributed to an upstream rule"}
	}

	// Only a fixture whose diagnostics belong to exactly one wired rule can be scored against that
	// rule. A mixed fixture would be judged against one rule while expecting another's findings,
	// which manufactures a failure neither rule earned.
	if len(rules) != 1 {
		return react_conformance.Result{}, &react_conformance.ErrUnsupported{Reason: "fixture spans more than one upstream rule, so no single rule can be scored against it"}
	}
	subject, wired := Rules[rules[0]]
	if !wired {
		return react_conformance.Result{}, &react_conformance.ErrUnsupported{Reason: "no rule engine wired for upstream rule " + rules[0]}
	}
	return Analyze(subject, fixture, e.t.TempDir())
}

// TestAggregateWithTheRuleEngineWired is the whole-corpus score with real rules running.
//
// The sibling package's `TestAggregatedScoreAgainstReactsOwnFixtures` reports the same partition
// with `NothingImplemented`, so every addressable fixture lands in `not scored: no rule engine run`.
// This is that report with the engine supplied, and the difference between the two is the entire
// contribution of this file.
//
// The assertions are two-sided on purpose. A pass floor alone is satisfied by a harness that cannot
// see, and a failure count alone is satisfied by one that is broken; requiring BOTH to be non-zero,
// plus the partition check, is what makes the number a measurement.
func TestAggregateWithTheRuleEngineWired(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds a TypeScript program per fixture")
	}
	fixtures, err := react_conformance.Load(fixtureRoot)
	if err != nil {
		t.Fatalf("loading the corpus: %v", err)
	}

	report, err := react_conformance.Aggregate(fixtures, engineImplementation{t: t})
	if err != nil {
		t.Fatalf("aggregating: %v", err)
	}

	t.Logf("\n%s", report.Summary())
	for _, ruleName := range report.RuleNames() {
		if _, shipped := react_conformance.ShippedRules[ruleName]; !shipped {
			continue
		}
		t.Logf("  shipped rule %-22s %v", ruleName, report.ByRule()[ruleName])
	}

	if report.Counts[react_conformance.VerdictPassed] == 0 {
		t.Error("nothing passed, so the engine is not reaching any rule")
	}
	if report.Counts[react_conformance.VerdictFailed] == 0 {
		t.Error("nothing failed; a harness that cannot produce a failure is not measuring anything")
	}

	// The pinned numbers. Exact rather than floors, because the whole point of the wiring is that
	// this number can now be WRONG, and a floor cannot tell an improvement from a regression.
	//
	// Measured 2026-08-24 against facebook/react@bd6ea412c673.
	for _, want := range []struct {
		Verdict react_conformance.Verdict
		Count   int
	}{
		{react_conformance.VerdictPassed, 37},
		{react_conformance.VerdictFailed, 25},
		{react_conformance.VerdictStatedDivergence, 18},
		{react_conformance.VerdictUnresolvableTypes, 8},
		// The only one of the five that moved when the clean population was vendored, and the four
		// that did not are the reassurance: adding 70 fixtures nothing addresses must change the
		// exclusion count and leave every scored verdict alone.
		{react_conformance.VerdictFlowSyntax, react_conformance.ExpectedFlowFixtureCount},
	} {
		if got := report.Counts[want.Verdict]; got != want.Count {
			t.Errorf("%s = %d, want %d", want.Verdict, got, want.Count)
		}
	}

	// Nothing addressable may remain unscored. This is the assertion that would have failed before
	// this file existed, when all 178 sat in `not scored`, and it is the one that stops the engine
	// quietly ceasing to run.
	for _, upstream := range UpstreamNames() {
		for _, verdict := range report.Fixtures {
			if len(verdict.Rules) != 1 || verdict.Rules[0] != upstream {
				continue
			}
			if verdict.Verdict == react_conformance.VerdictNotScored {
				t.Errorf("%s is owned by wired rule %s and was not scored: %s", verdict.Name, upstream, verdict.Reason)
			}
		}
	}
}

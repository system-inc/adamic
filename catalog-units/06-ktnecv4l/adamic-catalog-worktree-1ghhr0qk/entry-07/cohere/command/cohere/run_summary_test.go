package main

import (
	"fmt"
	"testing"
)

// coveredSkipCases are the ruling of 2026-10-04 on a rule that skipped every file: a skip another check
// covers leaves nothing unchecked only when that check ran this run.
var coveredSkipCases = []struct {
	name   string
	skip   ruleSkip
	phases []phaseRecord
	want   int
}{
	{
		// ahra: noImplicitReturns is on, the types phase ran, and lint ran behind it.
		name: "covered, and its cover ran",
		skip: ruleSkip{Rule: "nexus/correctness-no-implicit-return", CoveredBy: "types"},
		phases: []phaseRecord{
			{Name: phaseFix, Outcome: outcomeRan}, {Name: phaseTypes, Outcome: outcomeRan}, {Name: phaseLint, Outcome: outcomeRan},
		},
		want: 0,
	},
	{
		name: "covered, under --lint, so the types phase did not run",
		skip: ruleSkip{Rule: "nexus/correctness-no-implicit-return", CoveredBy: "types"},
		phases: []phaseRecord{
			{Name: phaseFix, Outcome: outcomeSkipped}, {Name: phaseTypes, Outcome: outcomeSkipped, Detail: "not requested"},
			{Name: phaseLint, Outcome: outcomeRan},
		},
		want: 1,
	},
	{
		name: "covered, and the types phase bailed",
		skip: ruleSkip{Rule: "nexus/correctness-no-implicit-return", CoveredBy: "types"},
		phases: []phaseRecord{
			{Name: phaseFix, Outcome: outcomeRan}, {Name: phaseTypes, Outcome: outcomeRan},
			{Name: phaseLint, Outcome: outcomeNotReached, Detail: "types bailed"},
		},
		want: 1,
	},
	{
		name: "uncovered, on a run where every phase ran",
		skip: ruleSkip{Rule: "typescript/no-useless-default-assignment", Reason: "strictNullChecks is off"},
		phases: []phaseRecord{
			{Name: phaseFix, Outcome: outcomeRan}, {Name: phaseTypes, Outcome: outcomeRan}, {Name: phaseLint, Outcome: outcomeRan},
		},
		want: 1,
	},
}

func TestACoveredSkipIsAGapOnlyWhenItsCoverDidNotRun(t *testing.T) {
	t.Parallel()
	if wrong := wrongCoveredSkipVerdicts(runSummary.uncoveredSkips); len(wrong) > 0 {
		for _, failure := range wrong {
			t.Error(failure)
		}
	}

	// The mutant counts a skip as covered whenever it names a cover, ran or not. It must fail, or the cases
	// above would pass an implementation that never asked whether the types phase ran.
	ignoresWhetherTheCoverRan := func(summary runSummary) int {
		count := 0
		for _, skip := range summary.Skips {
			if skip.CoveredBy == "" {
				count++
			}
		}
		return count
	}
	if len(wrongCoveredSkipVerdicts(ignoresWhetherTheCoverRan)) == 0 {
		t.Error("an implementation that ignores whether the cover ran passed every case, so the cases do not test it")
	}
}

// The footer marks the gap the count decides, and only then.
func TestACoveredSkipIsMarkedOnlyAsAGap(t *testing.T) {
	t.Parallel()
	for _, testCase := range coveredSkipCases {
		summary := cleanSummary()
		summary.Phases, summary.Skips = testCase.phases, []ruleSkip{testCase.skip}
		summary.Gaps.RulesSkippingEverything = summary.uncoveredSkips()
		marked := carriesMarker(footer(summary, plain, footerOptions{}), "⚠ 1 rule skipped every file") == nil
		if marked != (testCase.want == 1) {
			t.Errorf("%s: marked %t, want %t", testCase.name, marked, testCase.want == 1)
		}
	}
}

func wrongCoveredSkipVerdicts(count func(runSummary) int) []string {
	var wrong []string
	for _, testCase := range coveredSkipCases {
		summary := runSummary{Phases: testCase.phases, Skips: []ruleSkip{testCase.skip}}
		if got := count(summary); got != testCase.want {
			wrong = append(wrong, fmt.Sprintf("%s: counted %d, want %d", testCase.name, got, testCase.want))
		}
	}
	return wrong
}

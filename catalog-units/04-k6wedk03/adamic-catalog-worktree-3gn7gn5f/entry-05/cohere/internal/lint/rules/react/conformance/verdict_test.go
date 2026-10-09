package react_conformance

import (
	"path/filepath"
	"strings"
	"testing"
)

// declineEverything is the implementation used to measure the corpus without a rule engine.
//
// It declines rather than returning an empty result, for the reason NothingImplemented documents:
// an empty result is the claim "I found no errors here", which is a wrong answer on a corpus where
// every fixture expects at least one.
type declineEverything struct{ reason string }

func (d declineEverything) Analyze(Fixture) (Result, error) {
	return Result{}, &ErrUnsupported{Reason: d.reason}
}

// TestAggregatedScoreAgainstReactsOwnFixtures is the number this package exists to produce.
//
// It is deliberately run against an implementation that answers nothing, and that is not a
// placeholder. Wiring cohere's rule engine in here would require building a TypeScript program per
// fixture — roughly a second each, so about five minutes for the corpus — inside a suite whose
// stated design property is that it costs 0.3 to 0.8 seconds and therefore never gets gated behind
// a tag. Paying that to watch 293 fixtures land in a decline bucket that is already decidable from
// the corpus alone would buy one number, `failed`, at the cost of the suite being run.
//
// What is asserted instead is every number that does NOT depend on running a rule: which fixtures
// any shipped rule could even be asked about, and why each of the others cannot be. That is the
// part that was never measured, and it is the part that decides whether a raw score would have
// lied.
//
// The headline, at the pinned sha: of 325 fixtures, 94 carry a diagnostic a rule cohere ships is
// responsible for. 196 belong entirely to rules cohere does not implement, and 35 need a Flow
// parser. Of the 94, eleven have been run against the real rule and every one landed in `stated
// divergence`; the remaining 83 are unscored. Every fixture has a named reason, counted here.
func TestAggregatedScoreAgainstReactsOwnFixtures(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	report, err := Aggregate(fixtures, declineEverything{reason: "scored without a rule engine; see the note on this test"})
	if err != nil {
		t.Fatalf("aggregating: %v", err)
	}

	if report.Considered != ExpectedFixtureCount {
		t.Errorf("considered %d, want %d", report.Considered, ExpectedFixtureCount)
	}

	// Flow count, unchanged from `TestFlowFixtureCountIsUpstreamsOwnTest`, restated here so a change in
	// the exclusion rule shows up in the aggregate rather than only in the loader's own test.
	if got := report.Counts[VerdictFlowSyntax]; got != ExpectedFlowFixtureCount {
		t.Errorf("flow-excluded %d, want %d", got, ExpectedFlowFixtureCount)
	}

	// 166 fixtures belong entirely to rules cohere does not ship: 101 of the error population, and
	// all 65 non-Flow fixtures of the clean one.
	//
	// The clean fixtures all land here for a reason that is correct rather than incidental.
	// `ShippedRules` does not list `preserve-manual-memoization`, because that rule's registration is
	// deliberately held: it reports on valid code, and a gate that cries wolf gets switched off. The
	// clean population exists precisely to measure that, so a corpus entry counted as addressable
	// while the rule that addresses it is unregistered would be the score claiming credit for a rule
	// nothing runs. When the registration lands, these move into the addressable set and this number
	// drops by 65.
	//
	// It remains the single biggest fact about the number and the one a headline score would bury:
	// a third of React's error corpus is preserve-manual-memoization, todo, invariant and friends,
	// none of which cohere has.
	if got := report.Counts[VerdictNoRuleShipped]; got != 166 {
		t.Errorf("no-rule-shipped %d, want 166", got)
	}

	// The addressable set: fixtures a rule cohere ships is responsible for. This is the real
	// denominator any future parity number is measured against, and it is 189 rather than the
	// corpus size. Unchanged by the clean population, which is the point of asserting it separately
	// from `Considered`: adding 70 fixtures nothing can be asked about must not move this number.
	answerable := report.Counts[VerdictPassed] + report.Counts[VerdictFailed] +
		report.Counts[VerdictUnresolvableTypes] + report.Counts[VerdictUnreachableFixture] +
		report.Counts[VerdictNotScored] + report.Counts[VerdictStatedDivergence]
	if answerable != 189 {
		t.Errorf("fixtures a shipped rule could be asked about = %d, want 189", answerable)
	}

	// This test scores `NothingImplemented`, so it measures the CATEGORISATION rather than any rule:
	// 18 fixtures carry a stated divergence, and every other addressable fixture lands in
	// `not scored` because the implementation handed in here declines everything.
	//
	// That is deliberate and it is why this package still depends on nothing in the rule engine.
	// The same partition WITH real rules running is
	// `react_conformance_score.TestAggregateWithTheRuleEngineWired`, and the two are meant to differ:
	// this one answers "what could be asked", that one answers "what did the rules say".
	if got := report.Counts[VerdictStatedDivergence]; got != 18 {
		t.Errorf("stated-divergence %d, want 18", got)
	}
	if got := report.Counts[VerdictNotScored]; got != 171 {
		t.Errorf("not-scored %d, want 171", got)
	}

	t.Logf("\n%s", report.Summary())
	for _, ruleName := range report.RuleNames() {
		if _, shipped := ShippedRules[ruleName]; !shipped {
			continue
		}
		t.Logf("  shipped rule %-22s %v", ruleName, report.ByRule()[ruleName])
	}
}

// TestNoRuleShippedIsTheCorpusRatherThanTheRules names what the 261 are.
//
// Asserting the breakdown rather than the total, because the total is a number anybody could argue
// with and the breakdown is a list anybody can check. Six categories are most of React's error
// corpus, and cohere implements none of them; that is the honest reason the parity number is small,
// and it is a different reason from "our rules are wrong".
func TestNoRuleShippedIsTheCorpusRatherThanTheRules(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	perRule := map[string]int{}
	for _, fixture := range fixtures {
		if fixture.RequiresFlow() {
			continue
		}
		rules, _ := fixture.Rules()
		anyShipped := false
		for _, ruleName := range rules {
			if _, found := ShippedRules[ruleName]; found {
				anyShipped = true
			}
		}
		if anyShipped {
			continue
		}
		for _, ruleName := range rules {
			perRule[ruleName]++
		}
	}

	// `immutability` (62) and `refs` (30) were the two largest rows here until both shipped on
	// 2026-08-24, and their removal is the measurement rather than a tidy-up: 92 flow-free fixtures
	// stopped being "we never wrote this" on the day the rules landed.
	//
	// `todo` moved 36 -> 30 in the same change, and the six are enumerated rather than retyped,
	// because a pinned number that is edited to match whatever the code now says has stopped being
	// a pin. Each is a fixture whose diagnostics span `todo` AND a now-shipped rule, so it leaves
	// the set of fixtures no shipped rule claims at all: `error.todo-reassign-const.js`,
	// `fault-tolerance/error.try-finally-and-mutation-of-props.js`,
	// `fault-tolerance/error.try-finally-and-ref-access.js`,
	// `fault-tolerance/error.try-finally-ref-access-and-mutation.js`,
	// `fault-tolerance/error.var-declaration-and-mutation-of-props.js`, and
	// `fault-tolerance/error.var-declaration-and-ref-access.js`.
	for _, want := range []struct {
		Rule  string
		Count int
	}{
		{"todo", 30},
		{"preserve-manual-memoization", 31},
		{"invariant", 15},
		{"memo-dependencies", 8},
	} {
		if got := perRule[want.Rule]; got != want.Count {
			t.Errorf("unshipped rule %s covers %d flow-free fixtures, want %d", want.Rule, got, want.Count)
		}
	}
}

// TestEveryCategoryIsReachable is the calibration, and it is the most important test in this file.
//
// The package it extends already carries `TestRunnerCanDetectAPass`, written because a runner that
// always reports zero satisfies every test that only checks it reports zero. The same failure
// applies one level up and is easier to miss: a report with seven named categories, six of which no
// input can ever reach, prints a rich-looking breakdown that is really a single number in costume.
//
// So each category is driven to non-empty with a constructed input, through the real `Classify`
// rather than by writing the constant into a map. A category that cannot be produced here is not a
// category and should be deleted rather than shipped as a label.
func TestEveryCategoryIsReachable(t *testing.T) {
	t.Parallel()
	globalsFixture := Fixture{
		Name:   "reachability.globals.js",
		Source: "function Component() { someGlobal = true; }",
		Expected: Expectation{Errors: []ExpectedError{
			{Heading: "Error", Message: "Cannot reassign variables declared outside of the component/hook", Line: 1},
		}},
	}

	for _, testCase := range []struct {
		Name    string
		Fixture Fixture
		Result  Result
		Err     error
		Want    Verdict
	}{
		{
			Name:    "passed",
			Fixture: globalsFixture,
			Result: Result{Errors: []ReportedError{
				{Heading: "Error", Message: "Cannot reassign variables declared outside of the component/hook", Line: 1},
			}},
			Want: VerdictPassed,
		},
		{
			Name:    "failed: reported nothing on an answerable fixture",
			Fixture: globalsFixture,
			Result:  Result{},
			Want:    VerdictFailed,
		},
		{
			Name:    "failed: reported the wrong line",
			Fixture: globalsFixture,
			Result: Result{Errors: []ReportedError{
				{Heading: "Error", Message: "Cannot reassign variables declared outside of the component/hook", Line: 9},
			}},
			Want: VerdictFailed,
		},
		{
			Name:    "not scored: the implementation declined",
			Fixture: globalsFixture,
			Err:     &ErrUnsupported{Reason: "no rule engine was run"},
			Want:    VerdictNotScored,
		},
		{
			// Reached when a diagnostic cannot be attributed to any upstream rule, which is the
			// honest answer for a fixture whose owning rule is the unknown.
			Name: "declined: unreachable fixture",
			Fixture: Fixture{
				Name:   "reachability.unattributed.js",
				Source: "function Component() {}",
				Expected: Expectation{Errors: []ExpectedError{
					{Heading: "Error", Message: "a message upstream does not emit", Line: 1},
				}},
			},
			Want: VerdictUnreachableFixture,
		},
		{
			// A type-aware rule, a hook call, and no react import. This is the shape 128 of the
			// corpus's 202 hook-using fixtures actually have.
			Name: "declined: unresolvable types",
			Fixture: Fixture{
				Name:   "reachability.setstate.js",
				Source: "function Component() { const [x, setX] = useState(0); setX(1); }",
				Expected: Expectation{Errors: []ExpectedError{
					{Heading: "Error", Message: "Cannot call setState during render", Line: 1},
				}},
			},
			Result: Result{},
			Want:   VerdictUnresolvableTypes,
		},
		{
			// `preserve-manual-memoization` rather than `immutability`, which this case used until
			// `immutability` shipped on 2026-08-24. The case needs a rule cohere genuinely does not
			// implement, so picking one that is being actively ported makes the calibration expire
			// the day the port lands — which is exactly what happened here.
			Name: "declined: no rule shipped",
			Fixture: Fixture{
				Name:   "reachability.preserve-manual-memoization.js",
				Source: "function Component(props) { return useMemo(() => props.x, []); }",
				Expected: Expectation{Errors: []ExpectedError{
					{Heading: "Compilation Skipped", Message: "Existing memoization could not be preserved", Line: 1},
				}},
			},
			Want: VerdictNoRuleShipped,
		},
		{
			Name: "excluded: flow syntax",
			Fixture: Fixture{
				Name:     "reachability.flow.js",
				Source:   "// @flow\nfunction Component() {}",
				Expected: Expectation{Errors: []ExpectedError{{Heading: "Error", Message: "anything", Line: 1}}},
			},
			Want: VerdictFlowSyntax,
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			got := Classify(testCase.Fixture, testCase.Result, testCase.Err)
			if got.Verdict != testCase.Want {
				t.Errorf("classified as %q, want %q (reason: %s)", got.Verdict, testCase.Want, got.Reason)
			}
			if got.Verdict != VerdictPassed && got.Reason == "" {
				t.Error("a non-pass with no reason")
			}
		})
	}
}

// TestStatedDivergenceCategoryCanHoldAnEntry is the seventh category's reachability proof.
//
// It is separated from the table above because the map it reads is empty at the pinned sha, for the
// reason written at `statedDivergences`: cohere's stated boundaries and React's error-named corpus
// do not currently intersect. An empty category is exactly the thing `TestEveryCategoryIsReachable`
// exists to distrust, so the wiring is proven directly by putting an entry in and taking it out
// again, rather than by asserting the constant exists.
// Not parallel: it adds entries to the package statedDivergences map, which Classify reads for every fixture,
// so a parallel test classifying at the same moment would race it and could see the entry.
func TestStatedDivergenceCategoryCanHoldAnEntry(t *testing.T) {
	const name = "reachability.divergence.js"

	statedDivergences[name] = StatedDivergence{
		Fixture:  name,
		Boundary: "TestStatedDivergenceCategoryCanHoldAnEntry",
		Reason:   "cohere deliberately does not carry this diagnostic kind",
	}
	defer delete(statedDivergences, name)

	verdict := Classify(Fixture{
		Name:     name,
		Source:   "function Component() {}",
		Expected: Expectation{Errors: []ExpectedError{{Heading: "Error", Message: "anything", Line: 1}}},
	}, Result{}, nil)

	if verdict.Verdict != VerdictStatedDivergence {
		t.Fatalf("classified as %q, want %q; the category is decorative rather than wired", verdict.Verdict, VerdictStatedDivergence)
	}
	if !strings.Contains(verdict.Reason, "TestStatedDivergenceCategoryCanHoldAnEntry") {
		t.Errorf("the reason does not name the boundary that holds it: %s", verdict.Reason)
	}

	// And it must beat the categories below it. A deliberate decision reclassified as a corpus
	// artifact would silently empty this category the moment a divergence landed on a fixture that
	// also lacked an import, which is the likeliest way for it to land at all.
	statedDivergences["reachability.divergence-typed.js"] = StatedDivergence{
		Fixture:  "reachability.divergence-typed.js",
		Boundary: "TestStatedDivergenceCategoryCanHoldAnEntry",
		Reason:   "deliberate",
	}
	defer delete(statedDivergences, "reachability.divergence-typed.js")

	shadowed := Classify(Fixture{
		Name:   "reachability.divergence-typed.js",
		Source: "function Component() { const [x, setX] = useState(0); setX(1); }",
		Expected: Expectation{Errors: []ExpectedError{
			{Heading: "Error", Message: "Cannot call setState during render", Line: 1},
		}},
	}, Result{}, nil)
	if shadowed.Verdict != VerdictStatedDivergence {
		t.Errorf("a stated divergence on an import-less fixture classified as %q; the decision must outrank the corpus artifact", shadowed.Verdict)
	}
}

// TestUnresolvableTypesIsNotAnExcuseForSyntaxRules keeps the most abusable category honest.
//
// `VerdictUnresolvableTypes` is the one category that could quietly absorb real failures: it fires
// on silence, and silence is what a broken rule produces. The thing stopping it is that only a
// type-aware rule can claim it. This proves that restriction is live, by handing the same silent
// result to a syntax-only rule's fixture and requiring a failure.
func TestUnresolvableTypesIsNotAnExcuseForSyntaxRules(t *testing.T) {
	t.Parallel()
	// `globals` is syntax-only and is not in TypeAwareRules.
	verdict := Classify(Fixture{
		Name:   "reachability.globals-no-import.js",
		Source: "function Component() { useThing(); someGlobal = true; }",
		Expected: Expectation{Errors: []ExpectedError{
			{Heading: "Error", Message: "Cannot reassign variables declared outside of the component/hook", Line: 1},
		}},
	}, Result{}, nil)

	if verdict.Verdict != VerdictFailed {
		t.Errorf("a syntax-only rule that found nothing classified as %q, want %q; the types excuse is leaking",
			verdict.Verdict, VerdictFailed)
	}
}

// TestReactImportDetectionDoesNotOverMatch guards the measurement that sizes the types category.
//
// The count that justifies `VerdictUnresolvableTypes` is "128 of 202 hook-using fixtures do not
// import react", and it is only as good as this predicate. A substring test for `react` would match
// `react-dom` and `./react-utils`, shrinking the category and moving those fixtures into `failed` —
// an error in the direction that looks more rigorous, which is the direction least likely to be
// questioned.
func TestReactImportDetectionDoesNotOverMatch(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		Source string
		Want   bool
	}{
		{"import {useState} from 'react';", true},
		{`import React from "react";`, true},
		{"const React = require('react');", true},
		{"import {render} from 'react-dom';", false},
		{"import {thing} from './react-utils';", false},
		{"import {render} from '@testing-library/react';", false},
		{"// react is great\nfunction Component() {}", false},
		{"function Component() { useState(0); }", false},
	} {
		if got := (Fixture{Source: testCase.Source}).ResolvesReactTypes(); got != testCase.Want {
			t.Errorf("ResolvesReactTypes(%q) = %v, want %v", testCase.Source, got, testCase.Want)
		}
	}
}

// TestCorpusTypeResolutionIsMeasured pins the numbers behind the types category.
//
// These come from the vendored corpus rather than from the brief, which carried oxc's figures (1 of
// 13 resolved, 33% of hook-using fixtures missing the import). React's own corpus is worse on this
// axis, not better, so a stub matters more here than the brief assumed.
func TestCorpusTypeResolutionIsMeasured(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	// Counted per population rather than over the corpus as a whole, because the two differ sharply
	// and a combined number would hide it: 63% of hook-using error fixtures omit the import against
	// 13% of the clean ones. Summing them reports 51%, a figure describing neither population and
	// moving with the mix rather than with anything about React.
	var importing, hookUsing, hookUsingWithoutImport int
	var cleanImporting, cleanHookUsing, cleanHookUsingWithoutImport int
	for _, fixture := range fixtures {
		base := filepath.Base(fixture.Name)
		errorNamed := strings.HasPrefix(base, "error.") || strings.HasPrefix(base, "todo.error.")

		resolves := fixture.ResolvesReactTypes()
		if errorNamed {
			if resolves {
				importing++
			}
			if fixture.CallsHook() {
				hookUsing++
				if !resolves {
					hookUsingWithoutImport++
				}
			}
			continue
		}
		if resolves {
			cleanImporting++
		}
		if fixture.CallsHook() {
			cleanHookUsing++
			if !resolves {
				cleanHookUsingWithoutImport++
			}
		}
	}

	if importing != 77 {
		t.Errorf("error fixtures importing react = %d, want 77", importing)
	}
	if hookUsing != 202 {
		t.Errorf("error fixtures calling a hook = %d, want 202", hookUsing)
	}
	if hookUsingWithoutImport != 128 {
		t.Errorf("hook-using error fixtures with no react import = %d, want 128", hookUsingWithoutImport)
	}

	if cleanImporting != 61 {
		t.Errorf("clean fixtures importing react = %d, want 61", cleanImporting)
	}
	if cleanHookUsing != 69 {
		t.Errorf("clean fixtures calling a hook = %d, want 69", cleanHookUsing)
	}
	if cleanHookUsingWithoutImport != 9 {
		t.Errorf("hook-using clean fixtures with no react import = %d, want 9", cleanHookUsingWithoutImport)
	}

	t.Logf("error fixtures: %d of %d hook-using (%.0f%%) never import react, so a type-aware rule reads `any`",
		hookUsingWithoutImport, hookUsing, 100*float64(hookUsingWithoutImport)/float64(hookUsing))
	t.Logf("clean fixtures: %d of %d hook-using (%.0f%%) never import react",
		cleanHookUsingWithoutImport, cleanHookUsing, 100*float64(cleanHookUsingWithoutImport)/float64(cleanHookUsing))
}

// TestReportCheckCatchesANonPartition proves the arithmetic guard fires.
func TestReportCheckCatchesANonPartition(t *testing.T) {
	t.Parallel()
	report := Report{
		Considered: 2,
		Counts:     map[Verdict]int{VerdictPassed: 1},
		Fixtures:   []FixtureVerdict{{Name: "a", Verdict: VerdictPassed}},
	}
	if err := report.Check(); err == nil {
		t.Fatal("Check accepted a report whose categories do not sum to the denominator")
	}

	unreasoned := Report{
		Considered: 1,
		Counts:     map[Verdict]int{VerdictFailed: 1},
		Fixtures:   []FixtureVerdict{{Name: "a", Verdict: VerdictFailed}},
	}
	if err := unreasoned.Check(); err == nil {
		t.Fatal("Check accepted a non-pass with no reason")
	}
}

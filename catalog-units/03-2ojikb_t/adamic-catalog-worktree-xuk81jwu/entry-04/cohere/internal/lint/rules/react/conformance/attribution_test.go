package react_conformance

import (
	"sort"
	"testing"
)

// TestEveryDiagnosticIsAttributed is the coverage assertion, and it is written as an exact zero.
//
// The failure it guards is the one this package was already built around, one level up: a map that
// covers most of the corpus produces a per-rule table that looks entirely reasonable. Earlier
// revisions of this map left 110, then 79, then 53 of the 429 unattributed, and at every one of
// those stages the per-rule breakdown read as a plausible result — the categories were the ones you
// would expect, in roughly the proportions you would expect. Nothing about a 376-of-429 map looks
// partial from its output.
//
// So the residual is asserted at zero rather than reported. A message this map has not seen means
// upstream added or reworded a diagnostic, and the correct response is to go read the emission site
// and add it, not to let the fixture fall out of whichever rule's denominator it belonged to.
func TestEveryDiagnosticIsAttributed(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	unattributed := map[string]int{}
	total := 0
	for _, fixture := range fixtures {
		for _, expectedError := range fixture.Expected.Errors {
			total++
			if _, found := CategoryForMessage(expectedError.Message); !found {
				unattributed[expectedError.Message]++
			}
		}
	}

	if total != 429 {
		t.Errorf("walked %d diagnostics, want 429; the attribution denominator moved", total)
	}
	if len(unattributed) != 0 {
		count := 0
		for message, occurrences := range unattributed {
			count += occurrences
			t.Errorf("no category for %q (%d occurrences); read its emission site upstream and add it", message, occurrences)
		}
		t.Errorf("%d of %d diagnostics unattributed", count, total)
	}
}

// TestAttributionReproducesTheGoldensOwnHeadings is the cross-check, and it is the reason the map
// above can be believed rather than merely inspected.
//
// Every golden records a heading independently of anything this package computes. Upstream's
// printer assigns exactly one heading per category over a switch closed by `assertExhaustive`
// (`CompilerError.ts:565`), transcribed here as HeadingForCategory. So running each attributed
// category through that switch and comparing to the heading actually printed is two measurements of
// the same quantity that were not derived from each other — the same shape as
// `TestDeclaredCountMatchesParsedErrors`, and the same reason it works.
//
// This check was not used to build the map. It is applied to a finished map, so a misattribution
// across a heading boundary — a Todo scored as an Immutability, an Invariant scored as a Refs —
// cannot survive it.
//
// What it cannot catch is stated rather than left implicit: twenty categories share the heading
// `Error`, so a swap between two of those is invisible here. This is a necessary check, not a
// sufficient one, and the only thing that makes those twenty safe is that each was read at its
// emission site.
func TestAttributionReproducesTheGoldensOwnHeadings(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	checked, mismatches := 0, 0
	for _, fixture := range fixtures {
		for _, expectedError := range fixture.Expected.Errors {
			category, found := CategoryForMessage(expectedError.Message)
			if !found {
				continue
			}
			heading, known := HeadingForCategory(category)
			if !known {
				t.Errorf("%s: category %q has no heading; the transcription of upstream's printer is incomplete", fixture.Name, category)
				continue
			}
			checked++
			if heading != expectedError.Heading {
				mismatches++
				t.Errorf("%s: attributed %q to category %s, whose heading is %q, but the golden printed %q",
					fixture.Name, expectedError.Message, category, heading, expectedError.Heading)
			}
		}
	}

	if checked != 429 {
		t.Errorf("cross-checked %d headings, want 429", checked)
	}
	if mismatches != 0 {
		t.Errorf("%d attributions disagree with the heading their golden printed", mismatches)
	}
}

// TestEveryCategoryMapsToARule keeps the two transcriptions from drifting apart.
//
// categoryByMessage and ruleByCategory were transcribed from different upstream functions. A
// category that appears in one and not the other is a transcription slip, and its symptom would be
// a fixture silently belonging to no rule — which reads as "no rule covers this" rather than as a
// defect.
func TestEveryCategoryMapsToARule(t *testing.T) {
	t.Parallel()
	for _, category := range categoryByMessage {
		if _, found := RuleForCategory(category); !found {
			t.Errorf("category %q has no rule name; ruleByCategory is missing an entry from getRuleForCategory", category)
		}
		if _, found := HeadingForCategory(category); !found {
			t.Errorf("category %q has no heading", category)
		}
	}
	for _, pattern := range categoryByMessagePrefix {
		if _, found := RuleForCategory(pattern.Category); !found {
			t.Errorf("prefix category %q has no rule name", pattern.Category)
		}
	}

	// 26 categories, matching the ErrorCategory enum at the pinned sha. Asserted so that an
	// upstream category added or removed is visible here rather than only where it is used.
	if len(ruleByCategory) != 26 {
		t.Errorf("ruleByCategory holds %d categories, want 26 at sha %s", len(ruleByCategory), UpstreamSha)
	}
	if len(headingByCategory) != 26 {
		t.Errorf("headingByCategory holds %d categories, want 26", len(headingByCategory))
	}
}

// TestAttributionIsAMeasurementNotAGuess pins the per-rule breakdown.
//
// These are the numbers a reader would otherwise have to take on trust, and several of them
// contradict what the task brief carried, so pinning them is what makes the contradiction durable
// rather than a claim in a report nobody re-runs.
//
// The four that moved, and why each is not a transcription slip:
//
//   - error-boundaries: brief 2, real 0. All three fixtures whose names say error-boundaries expect
//     a Todo from BuildHIR instead, because React's lowering aborts on a `try` with no `catch`
//     before the ErrorBoundaries validator runs. The two fixtures that do exercise it are not
//     error-named and live in a `## Logs` block, outside this corpus entirely.
//   - set-state-in-effect: brief 7, real 0. Its five reporting fixtures are all in `## Logs`.
//   - void-use-memo: brief 2, real 0, and static-components: brief 1, real 0. Same cause. Both
//     porters independently measured zero vendored goldens for their rule and were right.
//   - gating: brief 3, real 1. The gating/ subdirectory holds 3 error fixtures, but 2 expect a
//     Config diagnostic (`Could not parse dynamic gating`) rather than a Gating one. A
//     subdirectory-name join would have credited gating with all three.
//   - globals: brief 13, real 11. 15 Globals diagnostics spread across 11 fixtures, not 13.
//
// The brief was right about config (4), set-state-in-render (13), unsupported-syntax (1), and
// use-memo (7).
func TestAttributionIsAMeasurementNotAGuess(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	fixturesPerRule := map[string]int{}
	diagnosticsPerRule := map[string]int{}
	for _, fixture := range fixtures {
		rules, complete := fixture.Rules()
		if !complete {
			t.Errorf("%s: not every diagnostic attributed", fixture.Name)
		}
		for _, ruleName := range rules {
			fixturesPerRule[ruleName]++
		}
		for _, expectedError := range fixture.Expected.Errors {
			category, found := CategoryForMessage(expectedError.Message)
			if !found {
				continue
			}
			ruleName, _ := RuleForCategory(category)
			diagnosticsPerRule[ruleName]++
		}
	}

	for _, want := range []struct {
		Rule        string
		Fixtures    int
		Diagnostics int
	}{
		{"capitalized-calls", 3, 3},
		{"config", 4, 4},
		{"error-boundaries", 0, 0},
		{"exhaustive-effect-dependencies", 4, 13},
		{"gating", 1, 1},
		{"globals", 11, 15},
		{"hooks", 57, 74},
		{"immutability", 65, 80},
		{"incompatible-library", 3, 3},
		{"invariant", 18, 18},
		{"memo-dependencies", 14, 18},
		{"no-deriving-state-in-effects", 1, 1},
		{"preserve-manual-memoization", 33, 41},
		{"purity", 2, 6},
		{"refs", 42, 52},
		{"rule-suppression", 5, 7},
		{"set-state-in-effect", 0, 0},
		{"set-state-in-render", 13, 18},
		{"static-components", 0, 0},
		{"syntax", 3, 3},
		{"todo", 39, 64},
		{"unsupported-syntax", 1, 1},
		{"use-memo", 7, 7},
		{"void-use-memo", 0, 0},
	} {
		if got := fixturesPerRule[want.Rule]; got != want.Fixtures {
			t.Errorf("%s: %d fixtures, want %d", want.Rule, got, want.Fixtures)
		}
		if got := diagnosticsPerRule[want.Rule]; got != want.Diagnostics {
			t.Errorf("%s: %d diagnostics, want %d", want.Rule, got, want.Diagnostics)
		}
	}

	total := 0
	for _, count := range diagnosticsPerRule {
		total += count
	}
	if total != 429 {
		t.Errorf("per-rule diagnostics sum to %d, want 429", total)
	}

	names := make([]string, 0, len(fixturesPerRule))
	for name := range fixturesPerRule {
		names = append(names, name)
	}
	sort.Strings(names)
	t.Logf("rules with at least one error-named fixture: %d of 26 (%v)", len(names), names)
}

// TestRulesShippedInVerifyWithNoErrorNamedFixture states the corpus gap as an assertion.
//
// Four rules cohere ships cannot be scored against this corpus at all, and the reason is the
// corpus rather than the rules. Writing it down as a test rather than a comment means that if
// upstream later adds an error-named fixture for one of them, this fails and someone goes and
// scores it, instead of the rule staying permanently in a decline bucket nobody revisits.
func TestRulesShippedInVerifyWithNoErrorNamedFixture(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	covered := map[string]bool{}
	for _, fixture := range fixtures {
		rules, _ := fixture.Rules()
		for _, ruleName := range rules {
			covered[ruleName] = true
		}
	}

	for _, ruleName := range []string{"error-boundaries", "set-state-in-effect", "static-components", "void-use-memo"} {
		if covered[ruleName] {
			t.Errorf("%s now has an error-named fixture; it was measured at zero when this was written, so score it rather than leaving it declined", ruleName)
		}
	}
}

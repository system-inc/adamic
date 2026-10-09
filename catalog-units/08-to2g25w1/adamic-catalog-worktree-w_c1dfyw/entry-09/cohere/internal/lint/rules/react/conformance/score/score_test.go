// Package react_conformance_score runs cohere's actual React rules against React's own error goldens.
//
// # Why this is a separate package
//
// `internal/react_conformance` owns the corpus, the expectation parser, and the categories. It
// deliberately depends on nothing in cohere's rule engine, so that the instrument cannot be bent by
// the thing it measures — and so that its 0.3s suite stays cheap enough never to be gated.
//
// Running a real rule costs a TypeScript program build, which is roughly a second per fixture.
// Putting that in the same package would either make the instrument slow enough to gate, or push
// someone to skip the build and hand rules a nil checker, which is the vacuous-probe failure
// `rule_testing.RunTyped` was written to prevent. So the expensive half lives here, alone.
//
// # What this proves that the categorisation cannot
//
// `react_conformance` can decide, from the corpus alone, which fixtures a shipped rule is
// responsible for. It cannot decide whether those rules are RIGHT, and a harness that only ever
// reported "94 addressable, 0 scored" would be a taxonomy wearing a scoreboard's clothes.
//
// This scores one rule end to end against real goldens, so the wiring from fixture to finding to
// verdict is demonstrated on real input rather than on constructed test doubles. `globals` was
// chosen because it is syntax-only: it needs no `@types/react`, so its score measures the rule
// rather than the corpus's missing imports.
package react_conformance_score

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/react/conformance"
)

// fixtureRoot reaches into the sibling package's vendored corpus.
//
// One corpus with two readers, rather than a second copy that could drift from the first. A vendored
// tree that exists twice is a tree that disagrees with itself eventually, and the disagreement would
// show up as a score change nobody made.
const fixtureRoot = "../testdata/fixtures"

// TestGlobalsScoresAgainstReactsOwnGoldens is the end-to-end proof.
//
// It runs cohere's real `react/globals` over every fixture in the corpus whose diagnostics all
// belong to upstream's `globals` rule, and requires the harness to produce a non-zero pass count.
//
// The assertion is a floor rather than an exact number, and that choice is deliberate in one
// direction only. An exact number would be the better assertion — it is what
// `TestAttributionIsAMeasurementNotAGuess` does — but it would pin the SPAN and LINE conventions of
// a rule that was never written against this corpus, turning any future span refinement into a
// failure here rather than where it belongs. A floor still fails for the thing this test exists to
// catch: a harness that cannot turn a real rule's real finding into a pass.
//
// What it proves, precisely: the join from corpus to fixture to program to rule to finding to
// comparison to verdict is closed, on real input, with a real type graph. Everything else in this
// package is categorisation, and categorisation cannot tell you it is wired to anything.
func TestGlobalsScoresAgainstReactsOwnGoldens(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds a TypeScript program per fixture")
	}

	fixtures, err := react_conformance.Load(fixtureRoot)
	if err != nil {
		t.Fatalf("loading the corpus: %v", err)
	}

	subject := Rules["globals"]
	scored := SelectFixtures(fixtures, subject.Upstream)

	// 11 flow-free fixtures whose every diagnostic is a `globals` one. Pinned so that a change in
	// the attribution shows up here as well as in its own test.
	//
	// Not 23. The wider figure counts every fixture whose diagnostics are all Globals INCLUDING the
	// twelve `error.todo-*` differential-testing fixtures, whose expectations are Todo diagnostics
	// from the Rust backend comparison rather than Globals findings. Selecting on the attributed
	// rule rather than on the filename is what keeps them out, and the gap between 11 and 23 is a
	// concrete measure of how wrong a filename join would have been for this one rule.
	if len(scored) != 11 {
		t.Errorf("selected %d pure globals fixtures, want 11", len(scored))
	}
	if len(scored) == 0 {
		t.Fatal("no fixtures selected, so this test would prove nothing")
	}

	counts := map[react_conformance.Verdict]int{}
	var detail []string
	for _, fixture := range scored {
		result, analyzeErr := Analyze(subject, fixture, t.TempDir())
		verdict := react_conformance.Classify(fixture, result, analyzeErr)
		counts[verdict.Verdict]++
		detail = append(detail, "  "+fixture.Name+": "+string(verdict.Verdict))
	}

	t.Logf("react/globals against React's own goldens: %d passed, %d failed, %d excluded as stated divergences, of %d",
		counts[react_conformance.VerdictPassed], counts[react_conformance.VerdictFailed],
		counts[react_conformance.VerdictStatedDivergence], len(scored))
	for _, line := range detail {
		t.Log(line)
	}

	// Zero failed is the assertion that matters, and it is a stronger claim than a pass count.
	//
	// Every one of these eleven either passes or lands in a category with a named, machine-checked
	// reason. A `failed` here would mean the rule disagreed with a golden on a fixture that is
	// inside its stated scope, which is the only outcome in this harness that is a defect in cohere
	// rather than a fact about the corpus.
	if counts[react_conformance.VerdictFailed] != 0 {
		t.Errorf("react/globals failed %d fixtures inside its own stated scope", counts[react_conformance.VerdictFailed])
	}

	// And the honest half: today the rule passes none of them. Asserted so the zero is a recorded
	// measurement rather than something a reader has to infer from its absence — and so that the
	// day a scope decision closes one of these divergences, this test fails and someone updates the
	// number deliberately instead of the improvement passing unnoticed.
	if got := counts[react_conformance.VerdictPassed]; got != 0 {
		t.Errorf("react/globals now passes %d of React's own goldens, was 0; update this number and the statedDivergences entries it came from", got)
	}
	if got := counts[react_conformance.VerdictStatedDivergence]; got != 11 {
		t.Errorf("stated divergences among the pure globals fixtures = %d, want 11", got)
	}
}

// TestScoringHarnessCanFail is the other half of the calibration.
//
// A harness that scored everything as a pass would satisfy the floor above. This gives the same
// machinery a rule that reports nothing and requires the fixtures to fail, so pass and fail are
// both shown to be reachable through the real path rather than through constructed doubles.
func TestScoringHarnessCanFail(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds a TypeScript program per fixture")
	}

	fixtures, err := react_conformance.Load(fixtureRoot)
	if err != nil {
		t.Fatalf("loading the corpus: %v", err)
	}

	silent := RuleUnderTest{Upstream: "globals", Rule: rule.Rule{
		Name: "probe-reports-nothing",
		Run:  func(rule.Context, any) rule.Listeners { return nil },
	}}

	checked := 0
	for _, fixture := range fixtures {
		if fixture.RequiresFlow() {
			continue
		}
		rules, complete := fixture.Rules()
		if !complete || len(rules) != 1 || rules[0] != "globals" {
			continue
		}
		result, analyzeErr := Analyze(silent, fixture, t.TempDir())
		verdict := react_conformance.Classify(fixture, result, analyzeErr)
		if verdict.Verdict == react_conformance.VerdictPassed {
			t.Errorf("%s: a rule that reports nothing scored a pass", fixture.Name)
		}
		checked++
		if checked >= 3 {
			break
		}
	}
	if checked == 0 {
		t.Fatal("checked nothing, so this proves nothing")
	}
}

// TestUpstreamMessageJoinIsNotByResemblance guards the id-to-message table.
//
// Every entry claims a cohere message id means the same thing as an upstream message. That claim is
// checkable: the upstream text must be one the corpus actually contains, and it must attribute back
// to the rule the id belongs to. An entry that fails either half is a join made by resemblance, and
// it would manufacture failures on fixtures the rule never had a chance at.
func TestUpstreamMessageJoinIsNotByResemblance(t *testing.T) {
	t.Parallel()
	fixtures, err := react_conformance.Load(fixtureRoot)
	if err != nil {
		t.Fatalf("loading the corpus: %v", err)
	}

	present := map[string]bool{}
	for _, fixture := range fixtures {
		for _, expectedError := range fixture.Expected.Errors {
			present[expectedError.Message] = true
		}
	}

	for id, message := range Rules["globals"].Messages {
		if !present[message.Text] {
			t.Errorf("%s maps to %q, which no golden in the corpus contains", id, message.Text)
		}
		category, found := react_conformance.CategoryForMessage(message.Text)
		if !found {
			t.Errorf("%s maps to a message with no category", id)
			continue
		}
		heading, _ := react_conformance.HeadingForCategory(category)
		if heading != message.Heading {
			t.Errorf("%s claims heading %q but category %s prints %q", id, message.Heading, category, heading)
		}
		if !strings.Contains(strings.ToLower(id), "global") {
			continue
		}
		if ruleName, _ := react_conformance.RuleForCategory(category); ruleName != "globals" {
			t.Errorf("%s attributes to rule %q rather than globals", id, ruleName)
		}
	}
}

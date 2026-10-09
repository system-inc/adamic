package react_conformance_score

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rules/react/conformance"
)

// TestWiredRulesAgainstReactsOwnGoldens is the measurement: every wired rule over every fixture
// attributed to it, categorised, with each failure named.
//
// It asserts nothing about the totals yet; it is the instrument that produces the numbers the
// pinned assertions are then written from. A run that named no fixtures would be the tell that the
// selection is empty rather than the rules being perfect.
func TestWiredRulesAgainstReactsOwnGoldens(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds a TypeScript program per fixture")
	}
	fixtures, err := react_conformance.Load(fixtureRoot)
	if err != nil {
		t.Fatalf("loading the corpus: %v", err)
	}

	grand := map[react_conformance.Verdict]int{}
	for _, upstream := range UpstreamNames() {
		subject := Rules[upstream]
		scored := SelectFixtures(fixtures, upstream)
		counts := map[react_conformance.Verdict]int{}
		var failures []string
		for _, fixture := range scored {
			result, analyzeErr := Analyze(subject, fixture, t.TempDir())
			verdict := react_conformance.Classify(fixture, result, analyzeErr)
			counts[verdict.Verdict]++
			grand[verdict.Verdict]++
			if verdict.Verdict == react_conformance.VerdictFailed {
				failures = append(failures, "      "+fixture.Name+" :: "+verdict.Reason)
			}
		}
		t.Logf("%-20s selected %3d  passed %3d  failed %3d  divergence %2d  notscored %2d  unresolvable %2d",
			upstream, len(scored),
			counts[react_conformance.VerdictPassed],
			counts[react_conformance.VerdictFailed],
			counts[react_conformance.VerdictStatedDivergence],
			counts[react_conformance.VerdictNotScored],
			counts[react_conformance.VerdictUnresolvableTypes])
		for _, f := range failures {
			t.Log(f)
		}
	}
	t.Logf("TOTAL passed %d failed %d divergence %d notscored %d unresolvable %d",
		grand[react_conformance.VerdictPassed], grand[react_conformance.VerdictFailed],
		grand[react_conformance.VerdictStatedDivergence], grand[react_conformance.VerdictNotScored],
		grand[react_conformance.VerdictUnresolvableTypes])
}

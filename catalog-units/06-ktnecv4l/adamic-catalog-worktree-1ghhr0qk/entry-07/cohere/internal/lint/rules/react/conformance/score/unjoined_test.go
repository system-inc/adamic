package react_conformance_score

import (
	"sort"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rules/react/conformance"
)

// TestUnjoinedMessageIdsAreEnumerated measures what the message join DROPS.
//
// `Analyze` skips a finding whose message id is not in the rule's table, which is the right default
// — a wrong join manufactures a confident false failure — but it is silent, and silence about
// dropped findings is exactly what this harness exists to refuse. A rule could find the right thing
// under an unjoined id and score as a miss, and nothing would say so.
//
// This enumerates the drops instead. It exists because a mutation making unjoined ids map to a
// default message SURVIVED the whole suite: nothing could see the join's behaviour on an id outside
// the table, in either direction.
func TestUnjoinedMessageIdsAreEnumerated(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds a TypeScript program per fixture")
	}
	fixtures, err := react_conformance.Load(fixtureRoot)
	if err != nil {
		t.Fatalf("loading the corpus: %v", err)
	}

	// The ids the join is EXPECTED to carry, written as literals rather than read from
	// `Rules[...].Messages`.
	//
	// This is the whole reason this test works. The first version compared each observed id against
	// the live table, so deleting an entry removed it from the join AND from the comparison at the
	// same time and the two always agreed: a control mutation dropping a real, exercised entry
	// SURVIVED twice. Both sides moved together, which reads exactly like a correct join. Assert
	// against a literal you typed, never against the constant the code under test reports with.
	joinedIds := map[string]bool{
		"globals :: globalReassignment":                     true,
		"hooks :: rulesOfHooksConditional":                  true,
		"hooks :: rulesOfHooksLoop":                         true,
		"hooks :: rulesOfHooksCallback":                     true,
		"hooks :: rulesOfHooksNotComponent":                 true,
		"hooks :: rulesOfHooksTopLevel":                     true,
		"hooks :: rulesOfHooksClassComponent":               true,
		"hooks :: rulesOfHooksAsync":                        true,
		"set-state-in-render :: setStateInRender":           true,
		"set-state-in-render :: setStateInUseMemo":          true,
		"use-memo :: useMemoCallbackNotInline":              true,
		"use-memo :: useMemoDependencyListNotArrayLiteral":  true,
		"use-memo :: useMemoCallbackHasParameters":          true,
		"use-memo :: useMemoCallbackAsyncOrGenerator":       true,
		"use-memo :: useMemoCallbackReassignsOuterVariable": true,
		"config :: invalidTypeConfiguration":                true,
		"gating :: invalidGatingDirective":                  true,
		"unsupported-syntax :: unsupportedEval":             true,
	}

	dropped := map[string]int{}
	observedIds := map[string]int{}
	seen := 0
	for _, upstream := range UpstreamNames() {
		subject := Rules[upstream]
		// Widen the table to catch everything, then compare against the real one.
		for _, fixture := range SelectFixtures(fixtures, upstream) {
			raw, err := analyzeRaw(subject, fixture, t.TempDir())
			if err != nil {
				continue
			}
			for _, id := range raw {
				seen++
				observedIds[upstream+" :: "+id]++
				if !joinedIds[upstream+" :: "+id] {
					dropped[upstream+" :: "+id]++
				}
			}
		}
	}

	keys := make([]string, 0, len(dropped))
	for key := range dropped {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		t.Logf("  dropped %3d findings: %s", dropped[key], key)
	}
	for _, k := range func() []string {
		var o []string
		for k := range observedIds {
			o = append(o, k)
		}
		sort.Strings(o)
		return o
	}() {
		t.Logf("  OBSERVED %s x%d", k, observedIds[k])
	}
	t.Logf("%d raw findings observed; %d distinct message ids dropped by the join", seen, len(keys))

	// The control for the zero above. `0 dropped` and `0 observed` print the same digit, and a
	// probe that sees nothing satisfies the assertion vacuously. Measured: the wired rules emit
	// findings on their own fixtures, so a zero below means the probe never ran.
	if seen == 0 {
		t.Fatal("observed no findings at all, so the dropped-id count proves nothing")
	}

	// Pinned. A new unjoined id appearing means either a rule grew a message or the join went stale,
	// and both deserve a deliberate look rather than silent absorption into the failure count.
	if len(keys) != 0 {
		t.Errorf("the message join drops %d distinct ids; enumerate them above and either join or state each", len(keys))
	}
}

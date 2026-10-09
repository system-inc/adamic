package registry

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rules/react/conformance"
)

// TestShippedRuleNamesAreRegistered makes the second column of `ShippedRules` executable.
//
// # Why this test has to exist, and why it lives here
//
// `ShippedRules` maps an upstream rule name to the name cohere registers that rule under. Until
// 2026-08-24 every value read `react/<name>`, and not one of those strings was a registered rule:
// the react-compiler rules register BARE. The map's own doc comment asserted the opposite, claiming
// a rename on either side would be "a visible edit rather than a silent unjoin".
//
// It was silent for a reason worth stating, because it generalises past this map. `Classify` reads
// only the KEY — `_, found := ShippedRules[ruleName]` — so the value was never evaluated by
// anything. A field that no code path reads cannot be wrong in a way any test can see, and its
// presence makes the file LOOK checked. That is worse than omitting it, and it is the mechanism
// behind the open finding that the parity guard cannot tell a namespaced spelling from a bare one.
//
// This test is the fix: it evaluates every value against the live registry. It lives in the registry
// rather than beside the map because `react_conformance` deliberately depends on nothing in cohere's rule
// engine, and importing `registry` there would couple the instrument to the thing it measures. It left
// `react_conformance_score` because that was the only reason score's test binary linked the whole rule
// set, so every rule edit relinked it and reran its 5s of tests (#6gct10n).
//
// Scored by mutation: setting `hooks` back to `react/rules-of-hooks` SURVIVED the entire suite
// before this test existed, and so did corrupting an unwired rule's value. Both are caught now.
func TestShippedRuleNamesAreRegistered(t *testing.T) {
	t.Parallel()
	registered := map[string]bool{}
	var names []string
	for _, registeredRule := range All() {
		registered[registeredRule.Name] = true
		names = append(names, registeredRule.Name)
	}
	sort.Strings(names)

	if len(registered) == 0 {
		t.Fatal("the registry is empty, so this test would pass vacuously")
	}

	for upstream, cohereName := range react_conformance.ShippedRules {
		if registered[cohereName] {
			continue
		}
		// Name the near miss rather than only the absence. A namespaced spelling is the specific
		// error this test was written for, so saying so beats "not found" and sending the reader to
		// grep the registry.
		hint := ""
		if bare := strings.TrimPrefix(cohereName, "react/"); bare != cohereName && registered[bare] {
			hint = "; the rule registers BARE as " + bare + ", with no `react/` namespace"
		}
		t.Errorf("ShippedRules[%q] = %q, which no rule registers%s", upstream, cohereName, hint)
	}
}

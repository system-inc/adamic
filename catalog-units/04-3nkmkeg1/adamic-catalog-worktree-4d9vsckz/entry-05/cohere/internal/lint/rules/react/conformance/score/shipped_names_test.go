package react_conformance_score

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rules/react/conformance"
)

// TestEveryWiredRuleIsAShippedRule keeps the engine's table and the categorisation's table joined.
//
// A rule wired here but absent from `ShippedRules` would run and then have its fixtures classified
// as `no rule shipped`, so it would score nothing while appearing to work. The reverse is allowed
// and expected: `ShippedRules` is what cohere HAS, `Rules` is what this harness has WIRED, and the
// gap between them is the honest measure of what is left to do.
func TestEveryWiredRuleIsAShippedRule(t *testing.T) {
	t.Parallel()
	for _, upstream := range UpstreamNames() {
		if _, shipped := react_conformance.ShippedRules[upstream]; !shipped {
			t.Errorf("rule %q is wired in the engine but absent from ShippedRules, so its fixtures would be classified as unshipped", upstream)
		}
		if Rules[upstream].Rule.Name == "" {
			t.Errorf("rule %q is wired with no underlying cohere rule", upstream)
		}
		if len(Rules[upstream].Messages) == 0 {
			t.Errorf("rule %q is wired with an empty message join, so it can never report anything", upstream)
		}
	}
}

// TestWiredRuleNamesMatchTheRegisteredRule checks the engine wires the rule it claims to.
//
// `RuleUnderTest.Upstream` names an upstream rule and `.Rule` carries a cohere rule value; nothing
// structural stops those two disagreeing. This asserts the cohere rule's own `.Name` is the one
// `ShippedRules` says implements that upstream rule, so wiring `react.UseMemo` under the key
// `gating` fails here rather than producing a confidently wrong per-rule scoreboard.
func TestWiredRuleNamesMatchTheRegisteredRule(t *testing.T) {
	t.Parallel()
	for _, upstream := range UpstreamNames() {
		want := react_conformance.ShippedRules[upstream]
		if got := Rules[upstream].Rule.Name; got != want {
			t.Errorf("engine wires %q to cohere rule %q, but ShippedRules says %q implements it", upstream, got, want)
		}
	}
}

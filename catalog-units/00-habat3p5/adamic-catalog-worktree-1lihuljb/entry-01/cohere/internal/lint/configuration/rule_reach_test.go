package configuration

import (
	"reflect"
	"testing"
)

// The memo must answer exactly what a fresh scan of the registry answers, and must scan each key once.
// The pairs are the ones the doc on sameRuling names: a prefixed key that is one ruling with its bare
// spelling, and a core rule beside its typescript-eslint twin, which share a ruling while being two rules.
func TestRuleReachAnswersAsTheScanAndAsksEachKeyOnce(t *testing.T) {
	t.Parallel()
	registered := []string{"x", "no-invalid-this", "@typescript-eslint/no-invalid-this", "eqeqeq", "nexus/other"}
	scan := func(key string) map[string]bool {
		reached := map[string]bool{}
		for _, name := range registered {
			if KeyReachesRule(key, name) {
				reached[name] = true
			}
		}
		return reached
	}
	keys := []string{"x", "nexus/x", "no-invalid-this", "@typescript-eslint/no-invalid-this", "eqeqeq", "unknown", "other", "nexus/other"}

	reach := newRuleReach(registered)
	for round := range 2 {
		for _, left := range keys {
			if got := reach.of(left); !reflect.DeepEqual(got, scan(left)) {
				t.Fatalf("round %d: %q reaches %v by the memo and %v by a scan", round, left, got, scan(left))
			}
			for _, right := range keys {
				want := false
				for name := range scan(right) {
					want = want || scan(left)[name]
				}
				if got := reach.sameRuling(left, right); got != want {
					t.Errorf("sameRuling(%q, %q) = %v, a scan says %v", left, right, got, want)
				}
				every := true
				for name := range scan(right) {
					every = every && scan(left)[name]
				}
				if got := reach.reachesEvery(left, right); got != every {
					t.Errorf("reachesEvery(%q, %q) = %v, a scan says %v", left, right, got, every)
				}
			}
		}
	}
	if len(reach.byKey) != len(keys) {
		t.Fatalf("the memo holds %d keys after asking about %d, so a key was scanned under two names or missed", len(reach.byKey), len(keys))
	}
	if !reach.sameRuling("nexus/x", "x") || !reach.sameRuling("@typescript-eslint/no-invalid-this", "no-invalid-this") || reach.sameRuling("eqeqeq", "x") {
		t.Fatal("the documented cases do not hold")
	}
}

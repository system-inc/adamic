package lower

import (
	"os"
	"strings"
	"testing"
)

func TestConditionAssertionProofMutants(t *testing.T) {
	source, err := os.ReadFile("testdata/condition_assertions/proven.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct{ name, source string }{
		{"normal return on falsy", strings.Replace(string(source), `throw new Error("false condition")`, `return`, 1)},
		{"throw gated on assertion level", "let enabled = false;\n" + strings.Replace(string(source), "if (!value)", "if (enabled && !value)", 1)},
		{"throw cancelled by finally", strings.Replace(string(source), `throw new Error("false condition")`, `{ try { throw new Error("false condition"); } finally { return; } }`, 1)},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			if mutant.source == string(source) {
				t.Fatal("mutation changed nothing")
			}
			_, err := lowerSource(t, mutant.source)
			if err == nil || !strings.Contains(err.Error(), "type predicate") {
				t.Fatalf("unsound condition proof: %v", err)
			}
			checked, err := lowerTypeScriptAssertionSource(t, mutant.source)
			if err != nil {
				t.Fatal(err)
			}
			if checked.PredicateChecks.Checked != 1 || checked.PredicateChecks.Proven != 0 {
				t.Fatalf("mutant must get its condition check: %+v", checked.PredicateChecks)
			}
			t.Logf("proof mutant refused in .a and checked in .ts: %v", mutant.name)
		})
	}
}

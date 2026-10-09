package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Optional-presence reservation must retain the actual literal domain in a copied shape.
func TestCheckedWiderWritesReservedContract(t *testing.T) {
	for _, value := range []int{1, 2} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "reserved.ts")
			source := fmt.Sprintf("interface Slot { value?: 1 }\nconst empty = {};\nconst actual: Slot = { ...empty };\nconst wide: { value?: number } = actual;\nwide.value = %d;\nconsole.log(String(actual.value));\n", value)
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			expected := truth
			if value == 2 {
				expected = run{exitCode: 70, stderr: []byte("adamic: panic: write failed: wide.value expects 1 | undefined, got 2\n")}
			}
			native, binary := nativelyUncached(t, program)
			if d := disagreement(expected, native); d != "" {
				t.Fatalf("native: %s, %+v", d, native)
			}
			if d := disagreement(expected, onJavaScriptBackend(t, program)); d != "" {
				t.Fatal(d)
			}
			if value == 1 {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
				return
			}
			changed := false
			change := func(value ir.Expression) ir.Expression {
				literal, ok := value.(ir.ObjectLiteral)
				if !ok {
					return value
				}
				for i := range literal.Missing {
					contract := literal.Missing[i].Contract
					if contract != nil && len(contract.Allowed) > 0 {
						copy := *contract
						copy.Allowed = nil
						literal.Missing[i].Contract = &copy
						changed = true
					}
				}
				return literal
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), change)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), change)
			if !changed {
				t.Fatal("reserved contract mutant did not change a literal set")
			}
			mutant, binary := nativelyUncached(t, program)
			if d := disagreement(truth, mutant); d != "" {
				t.Fatalf("mutant must compile and match Node: %s, %+v", d, mutant)
			}
			if d := disagreement(truth, onJavaScriptBackend(t, program)); d != "" {
				t.Fatal(d)
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			if disagreement(expected, mutant) == "" {
				t.Fatal("reserved literal-set mutant survived")
			}
			t.Log("reserved literal-set mutant compiled and stored 2; the pinned misfit caught it")
		})
	}
}

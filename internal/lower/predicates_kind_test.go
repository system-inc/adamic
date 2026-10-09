package lower

import (
	"os"
	"strings"
	"testing"
)

func TestPredicateKindClaimsStayProven(t *testing.T) {
	for _, name := range []string{"enum", "alias", "optional"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("testdata/predicates_kind/" + name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			for _, probe := range []struct{ name, old, replacement string }{
				{"true lie", "=== SyntaxKind.Identifier", "!== SyntaxKind.Identifier"},
				{"false lie", "=== SyntaxKind.Identifier", "=== SyntaxKind.Identifier && false"},
			} {
				t.Run(probe.name, func(t *testing.T) {
					mutated := strings.Replace(string(source), probe.old, probe.replacement, 1)
					if mutated == string(source) {
						t.Fatal("body mutant changed nothing")
					}
					_, err = lowerSource(t, mutated)
					if err == nil || !strings.Contains(err.Error(), "type predicate") {
						t.Fatalf("unsound body admitted: %v", err)
					}
					t.Logf("%s body mutant refused: %v", probe.name, err)
				})
			}
		})
	}
}

// The source and optional target remain complete. Removing the old optional
// refusal is admission of the shared view, not a proof that extra is absent.
func TestPredicateKindOptionalTargetRetainsFields(t *testing.T) {
	source := `type A = { readonly kind: 'a' }; type B = { readonly kind: 'b' }; type Claimed = { readonly kind: 'a'; readonly extra?: number }; function assertA(x: A | B): asserts x is Claimed { if (x.kind !== 'a') { throw new Error('bad'); } }`
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	if !program.CheckedFields["extra"] {
		t.Fatal("optional target field was erased rather than retained in the checked view")
	}
}

func TestPredicateOptionalAliasesKeepUnsupportedCases(t *testing.T) {
	for _, source := range []string{
		`interface N {readonly kind:number}; interface T extends N {readonly kind:1;readonly text?:number}; function isT(x:N):x is T{return x.kind===1};function read(x:T){return x.text}`,
		`interface N {readonly kind:number}; interface T extends N {readonly kind:1;readonly text?:string}; function isT(x:N):x is T{return x.kind===1};function read(x:T|undefined){return x?.text}`,
	} {
		_, err := lowerSource(t, source)
		if err == nil || !strings.Contains(err.Error(), "a checked field alias requiring an optional, accessor, or representation conversion") {
			t.Fatalf("unsupported alias admitted: %v", err)
		}
	}
}

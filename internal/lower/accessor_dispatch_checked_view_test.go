package lower

import (
	"strings"
	"testing"
)

// A checked predicate's argument.kind read is a checked view read with no receiver type
// (predicates_proof.go), and both emitters check any read whose View is set. Where a class owns
// get kind(), finishAccessors would route that read through a dispatcher whose fallback is a
// plain load, so the read stops at the read instead (Compiler's must on #4gpbn7w's -3). Mutant:
// the guard back to `expression.View != "" && CheckedField(...)`, and this lowers.
func TestAccessorDispatchRefusesAPredicateCheckedKindRead(t *testing.T) {
	t.Parallel()
	source := `interface A { readonly kind: number }
interface B extends A { readonly kind: 1; readonly text: string }
class G implements A { get kind(): number { return 2; } }
function isB(n: A): n is B { return n.kind === 1; }
const items: A[] = [new G(), { kind: 1 }];
for (const item of items) {
	if (isB(item)) {
		console.log(item.text.toUpperCase());
	} else {
		console.log("not b");
	}
}
`
	_, err := lowerSourceAs(t, source, ".ts")
	if err == nil || !strings.Contains(err.Error(), "a checked view read whose receiver may own an accessor of the same name") {
		t.Fatalf("want the checked view read stop, got %v", err)
	}
}

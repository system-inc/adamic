package lower

import (
	"strings"
	"testing"
)

// A checked predicate's argument.kind read is a checked view read with no receiver type
// (predicates_proof.go), and both emitters check any read whose View is set. Where a class owns
// get kind(), the read never reaches finishAccessors' dispatcher: refuseReachedContractGetters stops
// the program first, since kind is in the predicate's checked field contract (Loom job J-853dk25, on
// main 3a5c070f and at -3 alike). That stop is what keeps the dispatcher's own guard, View alone since
// Compiler's must on #4gpbn7w's -3, a defense with no live witness: if this stop ever moves, the guard is
// what holds, and this test going red is the sign to give the guard a witness of its own.
func TestAPredicateCheckedKindReadWithAGetterStopsAtItsContract(t *testing.T) {
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
	if err == nil || !strings.Contains(err.Error(), "a getter in a checked field contract") {
		t.Fatalf("want the contract getter stop, got %v", err)
	}
}

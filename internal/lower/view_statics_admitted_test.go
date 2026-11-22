package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A class constructor may stand for a viewed object only where no optional field the object names is a
// property Function gives every constructor (#s03ndgw), since a class that doesn't declare one reads
// Function's on Node where the view would read undefined: MaybeCalled's optional call keeps its contract
// false. A required field may have such a name: Inner's tag and count, Named's name and Called's call are
// each admitted, and the run time finds each only as a data static a class wrote (static_own_slot). The
// checker says which names those are (viewStaticsAdmitted). Mutant: viewStaticsAdmitted's optional test
// dropped, so every Function name is admitted, turns MaybeCalled true, and this goes red
// (scratch-k24/proof.sh).
func TestViewStaticsAdmittedAsksFunctionAndObject(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, `interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; }
interface Named { readonly tag: 'N'; readonly name: string; }
interface Called { readonly tag: 'C'; readonly call: number; }
interface MaybeCalled { readonly tag: 'M'; readonly call?: number; }
interface View extends Root { readonly value: number | Inner; readonly named: number | Named; readonly called: number | Called; readonly maybe: number | MaybeCalled; }
const raw = { kind: 'Holder' as const, value: 1, named: 2, called: 3, maybe: 4 };
const base: Root = raw;
const viewed = base as View;
console.log(typeof viewed.value + typeof viewed.named + typeof viewed.called + typeof viewed.maybe);
`, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"Inner": true, "Named": true, "Called": true, "MaybeCalled": false}
	found := map[string]bool{}
	for _, contract := range program.ViewContracts {
		admitted, named := want[contract.Name]
		if !named || contract.Kind != ir.ViewObject {
			continue
		}
		found[contract.Name] = true
		if contract.StaticsAdmitted != admitted {
			t.Errorf("%s: StaticsAdmitted %t, want %t (fields %#v)", contract.Name, contract.StaticsAdmitted, admitted, contract.Fields)
		}
	}
	for name := range want {
		if !found[name] {
			t.Errorf("no object contract named %s", name)
		}
	}
}

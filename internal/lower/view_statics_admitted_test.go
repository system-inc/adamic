package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A class constructor may stand for a viewed object only where no field the object names is a property
// Function or Object gives every constructor (#s03ndgw): Inner's tag and count can only be statics a
// class declares, so its contract admits one; Named's name and Called's call are a constructor's own
// whatever it declares, so theirs don't. The checker says which names those are (viewStaticsAdmitted).
// Maybe's optional extra admits no class either (#24kvb6k's Now): a member named like it, a static method, a
// static getter or an inherited static, would read as absent.
// Mutant: viewStaticsAdmitted's loop reading an empty field list turns Named and Called true, and this
// goes red (scratch-s03p1/proof.sh).
func TestViewStaticsAdmittedAsksFunctionAndObject(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, `interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; }
interface Named { readonly tag: 'N'; readonly name: string; }
interface Called { readonly tag: 'C'; readonly call: number; }
interface Maybe { readonly tag: 'M'; readonly extra?: number; }
interface View extends Root { readonly value: number | Inner; readonly named: number | Named; readonly called: number | Called; readonly maybe: number | Maybe; }
const raw = { kind: 'Holder' as const, value: 1, named: 2, called: 3, maybe: 4 };
const base: Root = raw;
const viewed = base as View;
console.log(typeof viewed.value + typeof viewed.named + typeof viewed.called + typeof viewed.maybe);
`, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"Inner": true, "Named": false, "Called": false, "Maybe": false}
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

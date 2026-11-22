package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A class constructor may stand for a viewed object only where no optional field the object names is a
// property Function gives every constructor (#s03ndgw), since a class that doesn't declare one reads
// Function's on Node where the view would read undefined: MaybeCalled's optional call keeps its contract
// false. A required field may have such a name: Inner's tag and count, Named's name and Called's call are
// each admitted, and the run time finds each only as a data static a class wrote (static_own_slot).
// Maybe's optional extra is admitted while no class declares a static method named extra, and isn't once
// Tool does (#24kvb6k): the run time keeps a static method with no name, so its walk would read extra as
// absent where Node reads the method. The checker says which names Function's are (viewStaticsAdmitted).
// Mutants (scratch-k24r/proof.sh): viewStaticsAdmitted's optional test dropped, so every Function name is
// admitted, turns MaybeCalled true; its static method rule dropped turns Maybe true beside Tool. Each goes red.
func TestViewStaticsAdmittedAsksFunctionAndObject(t *testing.T) {
	t.Parallel()
	const source = `interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; }
interface Named { readonly tag: 'N'; readonly name: string; }
interface Called { readonly tag: 'C'; readonly call: number; }
interface MaybeCalled { readonly tag: 'MC'; readonly call?: number; }
interface Maybe { readonly tag: 'M'; readonly extra?: number; }
interface View extends Root { readonly value: number | Inner; readonly named: number | Named; readonly called: number | Called; readonly maybeCalled: number | MaybeCalled; readonly maybe: number | Maybe; }
const raw = { kind: 'Holder' as const, value: 1, named: 2, called: 3, maybeCalled: 4, maybe: 5 };
const base: Root = raw;
const viewed = base as View;
console.log(typeof viewed.value + typeof viewed.named + typeof viewed.called + typeof viewed.maybeCalled + typeof viewed.maybe);
`
	expectStaticsAdmitted(t, "no static method", source, map[string]bool{"Inner": true, "Named": true, "Called": true, "MaybeCalled": false, "Maybe": true})
	expectStaticsAdmitted(t, "Tool's static extra", source+`class Tool {
	static extra(): number {
		return 9;
	}
}
console.log(String(Tool.extra()));
`, map[string]bool{"Inner": true, "Named": true, "Called": true, "MaybeCalled": false, "Maybe": false})
}

// expectStaticsAdmitted lowers source and holds each object contract want names to its StaticsAdmitted.
func expectStaticsAdmitted(t *testing.T, label, source string, want map[string]bool) {
	t.Helper()
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatalf("%s: %v", label, err)
	}
	found := map[string]bool{}
	for _, contract := range program.ViewContracts {
		admitted, named := want[contract.Name]
		if !named || contract.Kind != ir.ViewObject {
			continue
		}
		found[contract.Name] = true
		if contract.StaticsAdmitted != admitted {
			t.Errorf("%s: %s: StaticsAdmitted %t, want %t (fields %#v)", label, contract.Name, contract.StaticsAdmitted, admitted, contract.Fields)
		}
	}
	for name := range want {
		if !found[name] {
			t.Errorf("%s: no object contract named %s", label, name)
		}
	}
}

package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// An array may stand for a viewed object only where every field the object names is optional and none
// is a property an array has (#s03ndgw): Loose's a and b are no array's, so its contract admits one;
// Pushy's push and Stringy's toString are Array's and Object's, which the checker names, Lengthy's
// length and Indexed's 0 are an array's own whatever the library says, and Demanding's a is required,
// so none of theirs does (viewArraysAdmitted). Mutant: viewArraysAdmitted's loop reading an empty field
// list turns every one of them true, and this goes red (scratch-s03p2/proof.sh).
func TestViewArraysAdmittedAsksArrayAndObject(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, `interface Root { readonly kind: 'Holder'; }
interface Loose { readonly a?: number; readonly b?: string; }
interface Pushy { readonly push?: number; }
interface Stringy { readonly toString?: number; }
interface Lengthy { readonly length?: number; }
interface Indexed { readonly '0'?: number; }
interface Demanding { readonly a: number; }
interface View extends Root {
	readonly loose: number | Loose;
	readonly pushy: number | Pushy;
	readonly stringy: number | Stringy;
	readonly lengthy: number | Lengthy;
	readonly indexed: number | Indexed;
	readonly demanding: number | Demanding;
}
const raw = { kind: 'Holder' as const, loose: 1, pushy: 2, stringy: 3, lengthy: 4, indexed: 5, demanding: 6 };
const base: Root = raw;
const viewed = base as View;
console.log(typeof viewed.loose + typeof viewed.pushy + typeof viewed.stringy + typeof viewed.lengthy + typeof viewed.indexed + typeof viewed.demanding);
`, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"Loose": true, "Pushy": false, "Stringy": false, "Lengthy": false, "Indexed": false, "Demanding": false}
	found := map[string]bool{}
	for _, contract := range program.ViewContracts {
		admitted, named := want[contract.Name]
		if !named || contract.Kind != ir.ViewObject {
			continue
		}
		found[contract.Name] = true
		if contract.ArraysAdmitted != admitted {
			t.Errorf("%s: ArraysAdmitted %t, want %t (fields %#v)", contract.Name, contract.ArraysAdmitted, admitted, contract.Fields)
		}
	}
	for name := range want {
		if !found[name] {
			t.Errorf("no object contract named %s", name)
		}
	}
}

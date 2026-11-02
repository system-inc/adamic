package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// Language's v1 on f245bfe4: Element is a branded supertype with an optional name, and fresh is made
// without one, so its native shape has no slot for it. Element's required label is the required pin's.
const brandedWritePrelude = `interface Node { readonly kind: number; readonly pos: number; }
interface Declaration extends Node { _declarationBrand: undefined; }
interface Element extends Declaration { _elementBrand: undefined; name?: string; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
`

// A write through a branded supertype's checked view to a field it declares optional is NotYet, whatever
// the write's form, and whether or not the value holds the field, since its slot may be absent and the
// checked write can't add one (#c7w6j0y; Optimizer's ruling on Language's v1, pending Lower's word). A
// required field keeps its checked write and agrees with Node. Mutant: checkBrandedOptionalWrites off;
// with ADAMIC_BRANDED_OPTIONAL_WRITE_MUTANT=1 v1 must lower instead and its sanitized native must
// disagree with Node, which prints w.
func TestBrandedViewOptionalWriteIsNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"v1 assigned", `function rename(n: Node): void { if (isElement(n)) { n.name = "w"; console.log(String(n.name)); } }
const fresh: Node = { kind: 171, pos: 1, _declarationBrand: undefined, _elementBrand: undefined } as Node;
rename(fresh);`},
		{"v4 stored view", `function keep(n: Node): Element | undefined { return isElement(n) ? n : undefined; }
const fresh: Node = { kind: 171, pos: 1, _declarationBrand: undefined, _elementBrand: undefined } as Node;
const e = keep(fresh); if (e) { e.name = "w"; console.log(String(e.name)); }`},
		{"v2 present", `function rename(n: Node): void { if (isElement(n)) { n.name = "w"; console.log(String(n.name)); } }
const held: Node = { kind: 171, pos: 1, _declarationBrand: undefined, _elementBrand: undefined, name: "a" } as Node;
rename(held);`},
		{"nullish assigned", `function rename(n: Node): void { if (isElement(n)) { n.name ??= "w"; console.log(String(n.name)); } }
const fresh: Node = { kind: 171, pos: 1, _declarationBrand: undefined, _elementBrand: undefined } as Node;
rename(fresh);`},
		{"compound assigned", `function rename(n: Node): void { if (isElement(n)) { n.name += "x"; console.log(String(n.name)); } }
const held: Node = { kind: 171, pos: 1, _declarationBrand: undefined, _elementBrand: undefined, name: "a" } as Node;
rename(held);`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerTypeScriptAssertionSource(t, brandedWritePrelude+probe.source)
			if os.Getenv("ADAMIC_BRANDED_OPTIONAL_WRITE_MUTANT") == "1" {
				if probe.name != "v1 assigned" {
					t.Skip("the mutant is shown on v1")
				}
				if err != nil {
					t.Fatalf("the mutant must admit the program: %v", err)
				}
				code, out, errs := brandedNative(t, program)
				if code == 0 && out == "w\n" {
					t.Fatalf("the mutant's sanitized native agrees with Node: %q", out)
				}
				t.Logf("the mutant admits it: sanitized native exit %d, stdout %q, stderr %.300q where Node prints \"w\\n\"", code, out, errs)
				return
			}
			const what = "a write to optional field name through a checked predicate's view, whose slot may be absent"
			var pending *NotYet
			if !errors.As(err, &pending) || !strings.Contains(err.Error(), what) {
				t.Fatalf("want NotYet with %q, got %v", what, err)
			}
		})
	}
	t.Run("required field", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, brandedWritePrelude+`function relabel(n: Node): void { if (isElement(n)) { n.label = "z"; console.log(n.label); } }
const held: Node = { kind: 171, pos: 1, _declarationBrand: undefined, _elementBrand: undefined, label: "l" } as Node;
relabel(held);`, ".ts")
	})
}

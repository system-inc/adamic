package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// Language's v1 on f245bfe4: Element is a branded supertype with an optional name, and fresh is made
// without one, so its native shape has no slot for it. Element's count is the update's, and its
// required label the required pin's.
const viewOptionalWritePrelude = `interface Node { readonly kind: number; readonly pos: number; }
interface Declaration extends Node { _declarationBrand: undefined; }
interface Element extends Declaration { _elementBrand: undefined; name?: string; count?: number; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
`

// The reason checkViewOptionalWrites gives.
const viewOptionalWriteReason = "a write to optional field name through a checked view, whose slot a fixed shape can't add yet"

// A write through a checked view to a field the view declares optional is NotYet at the write, whatever
// its form, and whether or not the value holds the field, since its slot may be absent and the checked
// write can't add one (#c7w6j0y; Adamic, Lower and Compiler on Language's v1, Nov 2). A required field
// keeps its checked write and agrees with Node. Mutant: checkViewOptionalWrites off; with
// ADAMIC_VIEW_OPTIONAL_WRITE_MUTANT=1 v1 must lower instead and its sanitized native must disagree with
// Node, which prints w.
func TestViewOptionalWriteIsNotYet(t *testing.T) {
	t.Parallel()
	const fresh = `const fresh: Node = { kind: 171, pos: 1, _declarationBrand: undefined, _elementBrand: undefined } as Node;
`
	const held = `const held: Node = { kind: 171, pos: 1, _declarationBrand: undefined, _elementBrand: undefined, name: "a", count: 1 } as Node;
`
	for _, probe := range []struct{ name, source, want string }{
		{"v1 assigned", fresh + `function rename(n: Node): void { if (isElement(n)) { n.name = "w"; console.log(String(n.name)); } }
rename(fresh);`, viewOptionalWriteReason},
		{"v2 present", held + `function rename(n: Node): void { if (isElement(n)) { n.name = "w"; console.log(String(n.name)); } }
rename(held);`, viewOptionalWriteReason},
		{"v4 stored view", fresh + `function keep(n: Node): Element | undefined { return isElement(n) ? n : undefined; }
const e = keep(fresh); if (e) { e.name = "w"; console.log(String(e.name)); }`, viewOptionalWriteReason},
		{"compound", held + `function rename(n: Node): void { if (isElement(n)) { n.name += "x"; console.log(String(n.name)); } }
rename(held);`, viewOptionalWriteReason},
		{"nullish", fresh + `function rename(n: Node): void { if (isElement(n)) { n.name ??= "w"; console.log(String(n.name)); } }
rename(fresh);`, viewOptionalWriteReason},
		{"or", fresh + `function rename(n: Node): void { if (isElement(n)) { n.name ||= "w"; console.log(String(n.name)); } }
rename(fresh);`, viewOptionalWriteReason},
		{"increment", held + `function bump(n: Node): void { if (isElement(n)) { n.count!++; console.log(String(n.count)); } }
bump(held);`, viewOptionalWriteReason},
		{"delete", held + `function drop(n: Node): void { if (isElement(n)) { delete n.name; console.log(String(n.name)); } }
drop(held);`, ""},
		{"guarded", held + `function rename(n: Node): void { if (isElement(n)) { if (n.name !== undefined) { n.name = "w"; } console.log(String(n.name)); } }
rename(held);`, viewOptionalWriteReason},
		{"v3 cast", fresh + `function rename(n: Node): void { const e = n as Element; e.name = "w"; console.log(String(e.name)); }
rename(fresh);`, ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerTypeScriptAssertionSource(t, viewOptionalWritePrelude+probe.source)
			if os.Getenv("ADAMIC_VIEW_OPTIONAL_WRITE_MUTANT") == "1" {
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
			var pending *NotYet
			var refused *Refused
			if !errors.As(err, &pending) && !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want NotYet or Refused with %q, got %v", probe.want, err)
			}
			t.Logf("stops: %v", err)
		})
	}
	t.Run("required field", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, viewOptionalWritePrelude+`function relabel(n: Node): void { if (isElement(n)) { n.label = "z"; console.log(n.label); } }
const labelled: Node = { kind: 171, pos: 1, _declarationBrand: undefined, _elementBrand: undefined, label: "l" } as Node;
relabel(labelled);`, ".ts")
	})
}

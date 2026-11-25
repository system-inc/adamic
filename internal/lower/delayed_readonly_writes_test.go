package lower

import (
	"errors"
	"strings"
	"testing"
)

// A readonly view is never a write's slot (#3cee4gb, Language's J-2rzaksa, Adamic's ruling): a delayed
// mutator called on a readonly Ident[] seen as readonly Node[] wrote an Other where the Ident[] alias read
// an Ident, and native read past it. splice.call, fill.call, push.apply and reverse.call on a readonly
// receiver are refused at the call, as a direct write through a readonly view is. unshift.call keeps its
// own NotYet (no adapter). Mutant: delayedArrayWritesFit without its readonly test, and z01 to z03 lower
// again (z01 segfaults natively under ASan where Node throws).
func TestDelayedMutatorOnAReadonlyReceiverIsRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, method string }{
		{"z01 splice.call", `interface Node { readonly kind: number }
interface Ident extends Node { readonly kind: 1; readonly text: string }
interface Other extends Node { readonly kind: 2; readonly width: number }
const a: Ident = { kind: 1, text: "a" };
const b: Ident = { kind: 1, text: "b" };
const o: Other = { kind: 2, width: 7 };
const idents: readonly Ident[] = [a, b];
const nodes: readonly Node[] = idents;
Array.prototype.splice.call(nodes, 0, 1, o);
console.log(idents.map((i) => i.text.toUpperCase()).join(","));
`, "splice"},
		{"z02 fill.call", `interface Node { readonly kind: number }
interface Ident extends Node { readonly kind: 1; readonly text: string }
interface Other extends Node { readonly kind: 2; readonly width: number }
const a: Ident = { kind: 1, text: "a" };
const b: Ident = { kind: 1, text: "b" };
const o: Other = { kind: 2, width: 7 };
const idents: readonly Ident[] = [a, b];
const nodes: readonly Node[] = idents;
Array.prototype.fill.call(nodes, o);
console.log(idents.map((i) => i.text.toUpperCase()).join(","));
`, "fill"},
		{"z03 push.apply", `interface Node { readonly kind: number }
interface Ident extends Node { readonly kind: 1; readonly text: string }
interface Other extends Node { readonly kind: 2; readonly width: number }
const a: Ident = { kind: 1, text: "a" };
const b: Ident = { kind: 1, text: "b" };
const o: Other = { kind: 2, width: 7 };
const idents: readonly Ident[] = [a, b];
const nodes: readonly Node[] = idents;
Array.prototype.push.apply(nodes, [o]);
console.log(idents.map((i) => i.text.toUpperCase()).join(","));
`, "push"},
		{"reverse.call", `interface Node { readonly kind: number }
interface Ident extends Node { readonly kind: 1; readonly text: string }
interface Other extends Node { readonly kind: 2; readonly width: number }
const a: Ident = { kind: 1, text: "a" };
const b: Ident = { kind: 1, text: "b" };
const o: Other = { kind: 2, width: 7 };
const idents: readonly Ident[] = [a, b];
const nodes: readonly Node[] = idents;
Array.prototype.reverse.call(nodes);
console.log(idents.map((i) => i.text.toUpperCase()).join(","));
`, "reverse"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerTypeScriptAssertionSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "a delayed Array.prototype."+probe.method+" on a readonly array") {
				t.Fatalf("got %v, want the refusal of a delayed %s on a readonly array", err, probe.method)
			}
		})
	}
	// z04: unshift has no adapter, so it stops at its NotYet before the readonly test is asked.
	if _, err := lowerTypeScriptAssertionSource(t, `interface Node { readonly kind: number }
interface Ident extends Node { readonly kind: 1; readonly text: string }
interface Other extends Node { readonly kind: 2; readonly width: number }
const a: Ident = { kind: 1, text: "a" };
const b: Ident = { kind: 1, text: "b" };
const o: Other = { kind: 2, width: 7 };
const idents: readonly Ident[] = [a, b];
const nodes: readonly Node[] = idents;
Array.prototype.unshift.call(nodes, o);
console.log(idents.map((i) => i.text.toUpperCase()).join(","));
`); err == nil || !strings.Contains(err.Error(), "delayed Array.prototype.unshift") {
		t.Fatalf("z04: got %v, want the unshift NotYet", err)
	}
}

// The honest twin: the same splice.call on the mutable Ident[] the program owns, writing an Ident, lowers and
// agrees with Node.
func TestDelayedMutatorOnAMutableReceiverStillLowers(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeAs(t, `interface Node { readonly kind: number }
interface Ident extends Node { readonly kind: 1; readonly text: string }
interface Other extends Node { readonly kind: 2; readonly width: number }
const a: Ident = { kind: 1, text: "a" };
const b: Ident = { kind: 1, text: "b" };
const o: Other = { kind: 2, width: 7 };
const idents: Ident[] = [a, b];
Array.prototype.splice.call(idents, 0, 1, b);
console.log(idents.map((i) => i.text.toUpperCase()).join(","));
`, ".ts")
}

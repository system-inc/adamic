package lower

import (
	"errors"
	"strings"
	"testing"
)

const readonlyDelayedNodes = `interface Node { readonly kind: number }
interface Ident extends Node { readonly kind: 1; readonly text: string }
interface Other extends Node { readonly kind: 2; readonly width: number }
const a: Ident = { kind: 1, text: "a" };
const b: Ident = { kind: 1, text: "b" };
const o: Other = { kind: 2, width: 7 };
`

// A delayed Array mutator changes its receiver, and lib's Array.prototype declares no this, so call and apply
// hand it a readonly array as readily as a mutable one. A readonly array's element type is a covariant view:
// Language's z01 to z03 write an Other into a readonly Ident[] seen as readonly Node[], which the Ident[] alias
// then reads as an Ident, and native segfaulted on main 46a1954d where Node throws reading .text (#3cee4gb,
// J-2rzaksa). Her w rows write a readonly Node[] no narrower alias reads, which is still a write through a
// readonly view. A readonly view is never a write's slot (Adamic's ruling), so each is refused at the call,
// as are a readonly tuple and a NodeArray-shaped interface over ReadonlyArray. unshift has no adapter and keeps
// its NotYet. Mutant: delayedArrayReceiverWritable answering true, under which z01 to z03 and the w rows lower
// again and z01 dies natively.
func TestDelayedArrayMutatorsRefuseAReadonlyReceiver(t *testing.T) {
	t.Parallel()
	const aliased = readonlyDelayedNodes + "const idents: readonly Ident[] = [a, b];\nconst nodes: readonly Node[] = idents;\n"
	const read = "\nconsole.log(idents.map((i) => i.text.toUpperCase()).join(\",\"));\n"
	const list = readonlyDelayedNodes + "const list: readonly Node[] = [a, b];\n"
	const texts = "\nconsole.log(list.map((item) => (item.kind === 1 ? \"ident\" : \"other\")).join(\",\"));\n"
	for _, probe := range []struct{ name, source, want string }{
		{"z01 splice.call", aliased + "Array.prototype.splice.call(nodes, 0, 1, o);" + read, "a delayed Array splice that changes a readonly array"},
		{"z02 fill.call", aliased + "Array.prototype.fill.call(nodes, o);" + read, "a delayed Array fill that changes a readonly array"},
		{"z03 push.apply", aliased + "Array.prototype.push.apply(nodes, [o]);" + read, "a delayed Array push that changes a readonly array"},
		{"w01 splice.call", list + "Array.prototype.splice.call(list, 0, 1, o);" + texts, "a delayed Array splice that changes a readonly array"},
		{"w03 fill.call", list + "Array.prototype.fill.call(list, o);" + texts, "a delayed Array fill that changes a readonly array"},
		{"w04 reverse.call", list + "Array.prototype.reverse.call(list);" + texts, "a delayed Array reverse that changes a readonly array"},
		{"w07 push.apply", list + "Array.prototype.push.apply(list, [o]);" + texts, "a delayed Array push that changes a readonly array"},
		{"a readonly tuple", readonlyDelayedNodes + "const pair: readonly [Node, Node] = [a, b];\nArray.prototype.fill.call(pair, o);\nconsole.log(String(pair[0].kind));\n", "a delayed Array fill that changes a readonly array"},
		{"a method value held first", aliased + "const splice = Array.prototype.splice;\nsplice.call(nodes, 0, 1, o);" + read, "a delayed Array splice that changes a readonly array"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want Refused with %q, got %v", probe.want, err)
			}
		})
	}
}

// A NodeArray-shaped interface over ReadonlyArray, tsc's own shape, is no mutable Array, so a delayed mutator
// over it stops: refused by the readonly receiver, or NotYet earlier where its dense representation isn't
// proven. Either way it never lowers.
func TestDelayedArrayMutatorOverANodeArrayStops(t *testing.T) {
	t.Parallel()
	source := readonlyDelayedNodes + "interface NodeArray<T extends Node> extends ReadonlyArray<T> { readonly pos: number }\n" +
		"function grow(items: NodeArray<Node>): void { Array.prototype.push.call(items, o); }\n" +
		"const made: Node[] = [a, b];\nconst items = Object.assign(made, { pos: 0 }) as unknown as NodeArray<Node>;\ngrow(items);\nconsole.log(String(items.length));\n"
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	var pending *NotYet
	if !errors.As(err, &refused) && !errors.As(err, &pending) {
		t.Fatalf("want a stop, refused or NotYet, got %v", err)
	}
}

// The honest twins keep lowering and agree with Node natively under the sanitizers: a mutable Ident[] that
// splice.call writes an Ident into, a mutable Node[] seen through no readonly view, and a readonly array that
// a delayed method only reads (join), which leaves its receiver as it was. Mutant: the readonly test asked of
// every delayed method, under which the join row is refused.
func TestDelayedArrayMutatorsKeepAMutableReceiver(t *testing.T) {
	t.Parallel()
	source := readonlyDelayedNodes +
		"const idents: Ident[] = [a, b];\nArray.prototype.splice.call(idents, 0, 1, { kind: 1, text: \"c\" });\nconsole.log(idents.map((i) => i.text.toUpperCase()).join(\",\"));\n" +
		"const nodes: Node[] = [a, b];\nArray.prototype.push.apply(nodes, [o]);\nconsole.log(nodes.map((n) => String(n.kind)).join(\",\"));\n" +
		"const fixed: readonly string[] = [\"x\", \"y\"];\nconst joined: string = Array.prototype.join.call(fixed, \"|\");\nconsole.log(joined);\n"
	const stdout = "C,B\n1,1,2\nx|y\n"
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, out, stderr := brandedNative(t, program); code != 0 || out != stdout {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, out, stderr, stdout)
	}
}

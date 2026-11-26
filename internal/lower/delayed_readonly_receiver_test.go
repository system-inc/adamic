package lower

import (
	"errors"
	"strings"
	"testing"
)

const delayedReadonlyNodes = `interface Node { readonly kind: number }
interface Ident extends Node { readonly kind: 1; readonly text: string }
interface Other extends Node { readonly kind: 2; readonly width: number }
const a: Ident = { kind: 1, text: "a" };
const b: Ident = { kind: 1, text: "b" };
const o: Other = { kind: 2, width: 7 };
`

// The readonly test fails closed both ways (#3cee4gb, Compiler's ruling on the two heads). Which methods: the
// list is of what leaves the receiver as it was, so a mutator an adapter adds later refuses by default. Which
// receivers: only lib's mutable Array and a tuple that isn't readonly are writable, so an interface over
// ReadonlyArray that declares its own push, which a push-property test took as writable, is a readonly view
// and refuses. Language's w rows write a readonly Node[] no narrower alias reads, still a write through a
// readonly view. Mutant: main's push-property test and mutator list (#3cee4gb's first fix) put back. It
// survives every row here, honestly (J-ndrt32k): a readonly view that declares its own push is refused
// earlier, at the binding, by the method-parameter rule, a push declared as a property is a tsc error, and
// the mutator list names exactly the four adapters that exist. So these pins guard what no program can
// reach today. A fifth delayed adapter that changes its receiver, or a receiver shape tsc lets through
// with a push, would turn the mutant red on its row, and that row belongs here when it's added.
func TestDelayedMutatorReadonlyReceiverFailsClosed(t *testing.T) {
	t.Parallel()
	const list = delayedReadonlyNodes + "const list: readonly Node[] = [a, b];\n"
	const shown = "\nconsole.log(list.map((item) => (item.kind === 1 ? \"ident\" : \"other\")).join(\",\"));\n"
	for _, probe := range []struct{ name, source, method string }{
		{"w01 splice.call", list + "Array.prototype.splice.call(list, 0, 1, o);" + shown, "splice"},
		{"w03 fill.call", list + "Array.prototype.fill.call(list, o);" + shown, "fill"},
		{"w04 reverse.call", list + "Array.prototype.reverse.call(list);" + shown, "reverse"},
		{"w07 push.apply", list + "Array.prototype.push.apply(list, [o]);" + shown, "push"},
		{"a readonly member of a union", delayedReadonlyNodes + "const idents: readonly Ident[] = [a, b];\nconst mixed: Node[] | readonly Node[] = idents;\nArray.prototype.push.apply(mixed, [o]);\nconsole.log(idents.map((i) => i.text.toUpperCase()).join(\",\"));\n", "push"},
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
}

// An interface over ReadonlyArray that declares its own push is a covariant readonly view all the same: a
// readonly Ident[] seen through it as a Stack<Node> takes an Other the Ident[] alias reads. It never lowers:
// today it's refused at the binding by the method-parameter rule, before the delayed write, with the
// push-property test or without it (J-ndrt32k), and the readonly receiver test is the guard behind it.
func TestDelayedMutatorOnAReadonlyViewWithItsOwnPushStops(t *testing.T) {
	t.Parallel()
	source := delayedReadonlyNodes + "interface Stack<T> extends ReadonlyArray<T> { push(...items: T[]): number }\n" +
		"function grow(items: Stack<Node>): void { Array.prototype.push.apply(items, [o]); }\n" +
		"const idents: Ident[] = [a, b];\ngrow(idents);\nconsole.log(idents.map((i) => i.text.toUpperCase()).join(\",\"));\n"
	_, err := lowerTypeScriptAssertionSource(t, source)
	var refused *Refused
	var pending *NotYet
	if !errors.As(err, &refused) && !errors.As(err, &pending) {
		t.Fatalf("want a stop, refused or NotYet, got %v", err)
	}
	t.Logf("stops with %v", err)
}

// A method that leaves its receiver as it was keeps a readonly receiver: join.call over a readonly string[]
// lowers and agrees with Node. Mutant: the readonly test asked of every delayed method, under which it's
// refused.
func TestDelayedReaderOnAReadonlyReceiverStillLowers(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeAs(t, "const fixed: readonly string[] = [\"x\", \"y\"];\nconst joined: string = Array.prototype.join.call(fixed, \"|\");\nconsole.log(joined);\n", ".ts")
}

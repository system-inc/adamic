package lower

import (
	"strings"
	"testing"
)

// Rule G (#7fd51gk): a generic function seen as a generic slot is related as tsc relates it, the slot's
// type parameters held opaque and the source instantiated in their context. @types/node's own setTimeout,
// seen through sys.ts's timer slot, is TArgs instantiated as A, so its callback relates to the slot's.
func TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot(t *testing.T) {
	t.Parallel()
	source := `import { setTimeout } from 'node:timers';
interface Timers {
	setTimeout?<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A): unknown;
}
const timers: Timers = { setTimeout };
console.log(timers.setTimeout === undefined ? "none" : "some");
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, extension)
			if err == nil {
				return
			}
			for _, refusal := range []string{"seen as one taking", "falls back to a constraint holding any", "a generic method setTimeout"} {
				if strings.Contains(err.Error(), refusal) {
					t.Fatalf("the provider's relation to the slot stopped: %v", err)
				}
			}
		})
	}
}

// A generic method whose constraint differs from the view's but whose parameters are the view's is the
// same method once related (Language's i2 on #q6gs98r), so it lowers.
func TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers(t *testing.T) {
	t.Parallel()
	source := genericMethodAnimals + `interface Handler { handle<T>(x: Animal): string; }
class StringHandler { handle<T extends string>(x: Animal): string { return x.name; } }
const h: Handler = new StringHandler();
console.log(h.handle(new Cat()));
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, source, extension); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// A type parameter the slot doesn't fix falls back to its constraint, and one holding any would let the
// walk accept anything (rule G's condition 3): the provider's own NoInfer leaves T no candidate (a slot's
// NoInfer doesn't, since inference reads the slot's types as sources), so T is any[], and the provider
// would hand the slot's callback whatever its caller passed. Against a generic slot tsc refuses that itself
// (any[] isn't assignable to the slot's A), so the pin is the generic-to-plain branch, which tsc accepts. A generic function value stops
// earlier, at its own NotYet, so the pin is a class's method.
func TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet(t *testing.T) {
	t.Parallel()
	source := `interface Slot {
	run(callback: (...args: string[]) => void): void;
}
class Provider {
	run<T extends any[]>(callback: (...args: NoInfer<T>) => void): void {}
}
const slot: Slot = new Provider();
console.log("made");
`
	_, err := lowerSourceAs(t, source, ".ts")
	if err == nil || !strings.Contains(err.Error(), "falls back to a constraint holding any") {
		t.Fatalf("got %v, want the unfixed type parameter named as NotYet", err)
	}
}

// A generic target relates through rule G only where no written any meets a typed position (Adamic's
// ruling on #7fd51gk): tsc's forEachChildTable, its per-kind generic functions viewed through a cast as
// Record<SyntaxKind, ForEachChildFunction<any>> (parser.ts:1253), keeps main's refusal, since rule G would
// relate <T>(node: any) against <T>(node: QualifiedName) and pass through the program's own any.
func TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal(t *testing.T) {
	t.Parallel()
	source := `enum Kind { A = 1, B = 2 }
interface Node { readonly kind: Kind; }
interface NodeA extends Node { readonly kind: Kind.A; readonly a: number; }
interface NodeB extends Node { readonly kind: Kind.B; readonly b: string; }
type Nodes = NodeA | NodeB;
type Visit<TNode> = <T>(node: TNode, cb: (node: Node) => T | undefined) => T | undefined;
type Table = { [TNode in Nodes as TNode["kind"]]: Visit<TNode>; };
const table: Table = {
	[Kind.A]: function visitA<T>(node: NodeA, cb: (node: Node) => T | undefined): T | undefined { return node.a > 0 ? cb(node) : undefined; },
	[Kind.B]: function visitB<T>(node: NodeB, cb: (node: Node) => T | undefined): T | undefined { return node.b.length > 0 ? cb(node) : undefined; },
};
function each<T>(node: Node, cb: (node: Node) => T | undefined): T | undefined {
	const fn = (table as Record<Kind, Visit<any>>)[node.kind];
	return fn === undefined ? undefined : fn(node, cb);
}
const a: NodeA = { kind: Kind.A, a: 1 };
console.log(String(each(a, (node: Node): number => node.kind)));
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, extension)
			if err == nil {
				t.Fatal("lowered, but the cast relates the table's functions through a written any")
			}
			t.Logf("stops at: %v", err)
			if !strings.Contains(err.Error(), "13:14") && !strings.Contains(err.Error(), "13:13") {
				t.Fatalf("got %v, want the cast at 13:14 refused", err)
			}
		})
	}
}

package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// keepsNullishTypes is an open-kind interface and two kinded ones that extend it, the shape of tsc's Node and
// its members, for a cast that keeps undefined or null and claims a kind of the rest.
const keepsNullishTypes = "interface N { readonly kind: number } interface L extends N { readonly kind: 7; readonly m: string } interface M extends N { readonly kind: 8; readonly s: string } "

// A .ts cast that keeps its source's undefined or null and claims something of the rest (keepsNullishCast,
// #b5w3ycg) lowers to its operand held once in a local, then one Conditional: a nullish value is the local
// itself, and a present one is the inner cast's own check, a CheckedCast over the same local, so the tag is read
// only where a value is there. The inner road is each of castRoad's: an interface view, a kind union, and a
// union's tag; a source with no undefined of its own (binder.ts:2467's shape) takes it too. A source holding
// null tests null beside undefined, as the JavaScript backend holds an object's null as null. interfaceCast
// words its failure from the present types it's given; the other roads take the cast's own message.
// Mutant: the keeps-nullish road removed (the cast stops as before), and every row here stops with
// castKeepsNullish's NotYet.
func TestKeepsNullishCastLowersAsItsPresentRoad(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source, field, message string
		null                         bool
	}{
		{"interface view", keepsNullishTypes + "function f(x: N | undefined): L | undefined { return x as L | undefined; }", "kind",
			"cast failed: this N is not a L", false},
		{"kind union", keepsNullishTypes + "function f(x: N | undefined): L | M | undefined { return x as L | M | undefined; }", "kind",
			"cast failed: this N | undefined is not a L | M | undefined", false},
		{"no undefined in the source", keepsNullishTypes + "function f(x: N): L | undefined { return x as L | undefined; }", "kind",
			"cast failed: this N is not a L", false},
		{"null kept", keepsNullishTypes + "function f(x: N | null): L | null { return x as L | null; }", "kind",
			"cast failed: this N is not a L", true},
		// Held by absentArm's tests as a stop before this road: absentArm leaves a target that keeps undefined.
		{"a union's tag", "enum K { A = 1, B = 2 } type A = { readonly kind: K.A }; type B = { readonly kind: K.B }; function f(x: A | B | undefined): A | undefined { return x as A | undefined; }", "kind",
			"cast failed: this A | B | undefined is not a A | undefined", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatal(err)
			}
			composed := 0
			for _, function := range program.Functions {
				walk(function.Body, func(node any) bool {
					effects, ok := node.(ir.Effects)
					if !ok {
						return true
					}
					kept, ok := effects.Result.(ir.Conditional)
					if !ok {
						return true
					}
					composed++
					if len(effects.Body) != 1 {
						t.Fatalf("the cast's effects are %#v, want the operand's one Declare", effects.Body)
					}
					declare, declared := effects.Body[0].(ir.Declare)
					read, isRead := kept.WhenTrue.(ir.Read)
					if !declared || !isRead || declare.Local != read.Local || !program.Locals[read.Local].ExpressionAssigned {
						t.Fatalf("the nullish arm is %#v after %#v, want the local the operand is declared into", kept.WhenTrue, effects.Body[0])
					}
					test := kept.Condition
					if probe.null {
						either, isOr := test.(ir.Binary)
						if isNull, ok := either.Right.(ir.IsNull); !isOr || either.Operator != ir.Or || !ok || isNull.Value != ir.Expression(read) {
							t.Fatalf("the test is %#v, want undefined or null of the local", kept.Condition)
						}
						test = either.Left
					}
					if undefined, ok := test.(ir.IsUndefined); !ok || undefined.Value != ir.Expression(read) {
						t.Fatalf("the test is %#v, want undefined of the local", kept.Condition)
					}
					cast, ok := kept.WhenNot.(ir.CheckedCast)
					if !ok || cast.Value != ir.Expression(read) || cast.Field != probe.field {
						t.Fatalf("the present arm is %#v, want the inner cast's check of %s over the local", kept.WhenNot, probe.field)
					}
					if cast.Message != probe.message {
						t.Fatalf("message %q, want %q", cast.Message, probe.message)
					}
					return true
				})
			}
			if composed != 1 {
				t.Fatalf("got %d composed casts, want 1", composed)
			}
		})
	}
}

// What keepsNullishCast leaves as castProof judged it: a source's null cast to a target that keeps only undefined,
// and the reverse (an Object holds both as one null pointer, so the value would print as the target says where
// Node prints the other); a cast in .a, which keeps main's refusal word for word; and a present pair of classes,
// whose road is castProof's class branch, out of castRoad's reach. Each keeps its words. A present pair with no
// road of its own stops with that pair's cause, worded on the cast as written, since the undefined it keeps is
// no longer what stops it.
// Mutant: the keeps-nullish road removed (the cast stops as before) leaves the first four rows green and reds
// the last, which says castKeepsNullish's words again. keepsNullishCast's null test dropped reds "null kept as
// undefined", its undefined test "undefined kept as null", its .a test (checkedAssertionSource) the .a row, and
// its class test "present classes".
func TestKeepsNullishCastOutsideItsShapeStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, extension, words string }{
		{"null kept as undefined", keepsNullishTypes + "function f(x: N | null): L | undefined { return x as L | undefined; }", ".ts",
			"stage 0 can't lower a checked cast from N | null to L | undefined, which keeps undefined yet"},
		{"undefined kept as null", keepsNullishTypes + "function f(x: N | undefined): L | null { return x as L | null; }", ".ts",
			"stage 0 can't lower a checked cast from N | undefined to L | null, which keeps null yet"},
		{"in .a", keepsNullishTypes + "function f(x: N | undefined): L | undefined { return x as L | undefined; }", ".a",
			"Adamic 0.1 refuses a cast the runtime can't check; "},
		{"present classes", "class A { readonly n: number = 1; } class B extends A { readonly m: number = 2; } function f(x: A | undefined): B | undefined { return x as B | undefined; }", ".ts",
			"stage 0 can't lower a checked cast from A | undefined to B | undefined, which keeps undefined yet"},
		// C narrows A's field to a literal no view can prove, so the untagged arm (#axqgdck) keeps its stop. A cast to
		// A itself, a member, has the arm's checked view for a road and lowers.
		{"no present road", "interface A { readonly a: string } interface B { readonly b: string } interface C { readonly a: 'x' } function f(x: A | B | undefined): C | undefined { return x as C | undefined; }", ".ts",
			"stage 0 can't lower a checked view of C | undefined over the union A | B | undefined, whose members no literal tag tells apart yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			var refused *Refused
			var notYet *NotYet
			switch {
			case err == nil:
				t.Fatalf("lowered, want a stop saying %q", probe.words)
			case probe.extension == ".a" && !errors.As(err, &refused), probe.extension == ".ts" && !errors.As(err, &notYet):
				t.Fatalf("got %T: %v", err, err)
			case !strings.Contains(err.Error(), probe.words):
				t.Fatalf("want %q in the stop, got %v", probe.words, err)
			}
		})
	}
}

// The composed cast's view has the operand's local as its origin, a Read of a local the Effects declares, so what
// reaches the view is what the operand's value reaches through that Declare. A read through the view of a field
// whose contract stage 0 can't check yet (TestViewDiagnosticClassInstanceField's plain shape) stops at the read
// in the view's own words, as it does through `base as Viewed`. A write through the view stops sooner than a
// plain view's does: the optional-field write rule asks what reaches the receiver, and that reach doesn't yet
// forward the cast_operand local to its value, so the write stops at 7:2 on the unknown reach and the source-slot
// certificate (Rule C, TestSourceSlotPendingStopsOnlyWhereTheWriteIsLowered's "called on the viewed allocation")
// is never consulted. That fails closed, a stop and never an unchecked write; #e885ymf forwards the local and moves
// this row to the certificate's judgment.
// Mutant: the Declare's value left unforwarded for the read, which then lowers unchecked.
func TestKeepsNullishCastViewOriginIsTheOperandsLocal(t *testing.T) {
	t.Parallel()
	const read = `interface Root { readonly kind: 'Holder'; }
class Thing { readonly n = 5; }
interface Inner { readonly tag: 'I'; readonly thing: Thing; }
interface Viewed extends Root { readonly value: Inner; }
const raw = { kind: 'Holder' as const, value: { tag: 'I' as const, thing: new Thing() } };
const base: Root = raw;
const viewed = base as Viewed | undefined;
console.log(String(viewed!.value.thing.n));
`
	write := strings.Replace(sourceSlotPendingPrelude, "const view = node as NodeA;", "const view = node as NodeA | undefined;", 1)
	judgeSourceSlotPending(t, []sourceSlotPendingRow{
		{"a read through the view", read, "main.ts:8:20: stage 0 can't lower a class instance field thing of class Thing read through a checked view, which needs a check of the value's class at the read yet"},
		{"a write through the view", write + "setIt(view!, { id: 2 });\nconsole.log(String(node.flowNode === undefined));\n", "main.ts:7:2: stage 0 can't lower a write to optional field flowNode through a checked view, where what reaches its receiver isn't known yet"},
	})
}

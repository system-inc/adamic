package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// A slot's enum part is its numeric members alone, counted at the slot's own level, and it is an open enum
// slot only when they are exactly every member of the enum they come from, with no other numeric literal
// beside them (#snypn8c c2, Adamic's ruling). tsc's getPendingEmitKindWithSeen (builder.ts:788) takes
// `CompilerOptions | BuilderFileEmit`, and handing it a whole BuilderFileEmit was refused as an open enum into
// a literal slot: openNumericEnumType counted CompilerOptions among the slot's members, so the slot read as
// literal to freshOrWidened. A member, a strict subset, or a whole enum beside another enum's lone member stays
// a literal slot, refused in today's words, since rulings.md:247 holds only while every road from a number
// into an enum-literal tag slot is refused. Mutants (enum_slot_part.mutants.md): the object counted again,
// under which the 788 shape is refused; and any enum member making the slot open, under which a number goes
// into a member slot and beside another enum's member unrefused.

// The 788 shape, small: Emit is BuilderFileEmit, Options is CompilerOptions, and next hands its whole Emit
// parameter to emitOf, as 788 hands the pending map's emitKind to getPendingEmitKindWithSeen.
func TestWholeEnumIntoObjectOrEnumSlotLowers(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, `enum Emit { None = 0, Js = 1 << 0, Map = 1 << 1, Dts = 1 << 2, All = Js | Map | Dts }
interface Options {
	readonly declaration: boolean;
}
function emitOf(optionsOrKind: Options | Emit): Emit {
	return typeof optionsOrKind === 'number' ? optionsOrKind : optionsOrKind.declaration ? Emit.All : Emit.Js;
}
function next(kind: Emit): Emit {
	return emitOf(kind);
}
console.log(String(next(Emit.Map)));
console.log(String(next(Emit.Js | Emit.Dts)));
console.log(String(emitOf({ declaration: true })));
`)
}

// A number into a slot whose enum part is a whole enum lowers, beside an object as it does alone
// (rulings.md:243). maybe is builder.ts:1878's slot, with undefined beside the two.
func TestNumberIntoObjectOrWholeEnumSlotLowers(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, `enum Emit { None = 0, Js = 1 << 0, Map = 1 << 1, Dts = 1 << 2 }
interface Options {
	readonly declaration: boolean;
}
function emitOf(optionsOrKind: Options | Emit): number {
	return typeof optionsOrKind === 'number' ? optionsOrKind : optionsOrKind.declaration ? 4 : 1;
}
function maybe(value: Options | Emit | undefined): number {
	return value === undefined ? -1 : emitOf(value);
}
function widened(bits: number): Options | Emit {
	return bits;
}
const bits = Emit.Js | Emit.Dts;
console.log(String(emitOf(widened(bits))));
console.log(String(emitOf(bits * 8)));
console.log(String(maybe(bits | Emit.Map)));
console.log(String(maybe(undefined)));
console.log(String(emitOf({ declaration: false })));
`)
}

// The twin, a number into the whole enum alone, lowers as it always has (rulings.md:243).
func TestNumberIntoWholeEnumSlotLowers(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, `enum Emit { None = 0, Js = 1 << 0, Map = 1 << 1, Dts = 1 << 2 }
function alone(bits: number): Emit {
	return bits;
}
const bits = Emit.Js | Emit.Dts;
console.log(String(alone(bits * 8)));
`)
}

// memberSlotRefused is the refusal a number into one enum member literal gets, and fix its fix, today's words.
const (
	memberSlotRefused = "refuses an unproven value assigned to a numeric literal or enum member slot "
	memberSlotFix     = "; compare with this literal and return that constant, or widen the slot to the whole numeric enum or number (adamic/enum-literal)"
)

// A number into a slot whose enum part is one member, or a strict subset of three, stays refused beside an
// object, in the words it has alone.
func TestNumberIntoObjectOrMemberSlotStaysRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, slot string }{
		{"a member", "Options | E.A"},
		{"a strict subset", "Options | E.A | E.B"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := "enum E { A, B, C }\ninterface Options {\n\treadonly name: string;\n}\nfunction slot(value: number): " + probe.slot + " {\n\treturn value;\n}\n"
			_, err := lowerSource(t, source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), memberSlotRefused+"E.") || !strings.HasSuffix(err.Error(), memberSlotFix) {
				t.Fatalf("got %v, want a refusal naming a member of E with %q and ending %q", err, memberSlotRefused, memberSlotFix)
			}
		})
	}
}

// A whole enum beside another enum's lone member isn't every member of each enum its numeric members come
// from, so a number into it stays refused.
func TestNumberIntoWholeEnumBesideAnotherMemberStaysRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "enum E { A, B, C }\nenum F { A = 10, B = 11 }\nfunction slot(value: number): E | F.A {\n\treturn value;\n}\n")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), memberSlotRefused) || !strings.HasSuffix(err.Error(), memberSlotFix) {
		t.Fatalf("got %v, want a refusal with %q ending %q", err, memberSlotRefused, memberSlotFix)
	}
}

// A one-member numeric enum is never whole (Adamic's ruling on #x2cjvy9, condition 2): it's closed, so
// `Options | One` and `One` alone are member slots, and a number into either stays refused.
func TestNumberIntoOneMemberEnumSlotStaysRefused(t *testing.T) {
	t.Parallel()
	for name, slot := range map[string]string{"beside an object": "Options | One", "alone": "One"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, "enum One { A }\ninterface Options { readonly name: string }\nfunction slot(value: number): "+slot+" {\n\treturn value;\n}\n")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), memberSlotRefused) {
				t.Fatalf("got %v, want a refusal with %q", err, memberSlotRefused)
			}
		})
	}
}

// openEnumSlot asked of each shape above, beside openNumericEnumType, whose answer for a type that is exactly
// one enum's whole type doesn't move. The 788 slot reads as an enum's by its first member, so freshOrWidened
// asks openEnumSlot whether it's literal: that's how the shape above reaches the change.
func TestSlotEnumPartIsItsNumericMembers(t *testing.T) {
	t.Parallel()
	const source = `enum E { A, B, C }
enum F { A = 10, B = 11 }
interface Options {
	readonly name: string;
}
let whole: Options | E = E.A;
let wholeMaybe: Options | E | undefined = undefined;
let member: Options | E.A = E.A;
let subset: Options | E.A | E.B = E.A;
let beside: E | F.A = E.A;
let alone: E = E.A;
let worded: E | "x" = "x";
console.log(String(whole) + String(wholeMaybe) + String(member) + String(subset) + String(beside) + String(alone) + worded);
`
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	typeChecker, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{checker: typeChecker, program: program}
	declared := map[string]*ast.Node{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration {
			declared[node.Name().Text()] = node.Name()
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	for _, probe := range []struct {
		name       string
		slot, open bool
	}{
		{"whole", true, false},
		{"wholeMaybe", true, false},
		{"member", false, false},
		{"subset", false, false},
		{"beside", false, false},
		{"alone", true, true},
		{"worded", true, false},
	} {
		name, found := declared[probe.name]
		if !found {
			t.Fatalf("no declaration %s", probe.name)
		}
		proven := typeChecker.GetTypeAtLocation(name)
		if got := l.openEnumSlot(proven); got != probe.slot {
			t.Errorf("%s (%s): openEnumSlot %t, want %t", probe.name, typeChecker.TypeToString(proven), got, probe.slot)
		}
		if got := l.openNumericEnumType(proven); got != probe.open {
			t.Errorf("%s (%s): openNumericEnumType %t, want %t", probe.name, typeChecker.TypeToString(proven), got, probe.open)
		}
	}
	if whole := typeChecker.GetTypeAtLocation(declared["whole"]); !l.numericEnum(l.enumIdentity(whole)) {
		t.Errorf("whole (%s) doesn't read as a numeric enum's slot, so freshOrWidened never asks openEnumSlot of it", typeChecker.TypeToString(whole))
	}
}

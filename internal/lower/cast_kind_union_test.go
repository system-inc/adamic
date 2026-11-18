package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// kindUnionModel is tsc's shape in small: an open SyntaxKind on Node, and kinded interfaces below it, each
// with its own kind literal.
const kindUnionModel = `enum SyntaxKind { Identifier = 80, PrivateIdentifier = 81, MethodDeclaration = 174, PropertyAccessExpression = 211, ElementAccessExpression = 212, CallExpression = 213 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface Expression extends Node { readonly end: number }
interface Identifier extends Expression { readonly kind: SyntaxKind.Identifier; readonly text: string }
interface PropertyAccessExpression extends Expression { readonly kind: SyntaxKind.PropertyAccessExpression; readonly expression: Expression; readonly name: Identifier }
interface ElementAccessExpression extends Expression { readonly kind: SyntaxKind.ElementAccessExpression; readonly expression: Expression; readonly argumentExpression: Expression }
`

// A cast from an open-kind interface to a union of kinded interfaces (#b5w3ycg slice 4, Adamic's Nov 3 ruling)
// lowers in .ts to one tag test over exactly the target members' kinds, with the cast's own message, and a read
// through the result is checked, by the members' views. The program agrees with Node.
func TestKindUnionCastLowers(t *testing.T) {
	t.Parallel()
	source := kindUnionModel + `function target(node: Node): Expression {
	const access = node as PropertyAccessExpression | ElementAccessExpression;
	return access.expression;
}
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 1, end: 2, text: 'a' };
const property: PropertyAccessExpression = { kind: SyntaxKind.PropertyAccessExpression, pos: 0, end: 3, expression: id, name: id };
const element: ElementAccessExpression = { kind: SyntaxKind.ElementAccessExpression, pos: 4, end: 6, expression: property, argumentExpression: id };
console.log(String(target(property).pos) + ' ' + String(target(element).end));
`
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	casts, reads := 0, 0
	inspect := func(node any) bool {
		switch node := node.(type) {
		case ir.CheckedCast:
			casts++
			if node.Field != "kind" || len(node.Allowed) != 2 || node.FieldType != ir.Number || !node.CheckedFields {
				t.Errorf("the cast tests %q against %d values of type %v, checked fields %v; want kind against the two members' kinds", node.Field, len(node.Allowed), node.FieldType, node.CheckedFields)
			}
			// The checker orders a union's members by type id, so the kinds and the message follow its order.
			kinds := map[float64]bool{}
			for _, allowed := range node.Allowed {
				if constant, ok := allowed.(ir.NumberConstant); ok {
					kinds[constant.Value] = true
				}
			}
			if len(kinds) != 2 || !kinds[211] || !kinds[212] {
				t.Errorf("the cast allows %#v, want exactly the kinds 211 and 212", node.Allowed)
			}
			if !strings.HasPrefix(node.Message, "cast failed: this Node is not a ") || !strings.Contains(node.Message, "PropertyAccessExpression") || !strings.Contains(node.Message, "ElementAccessExpression") {
				t.Errorf("message %q, want the cast's own, naming both members", node.Message)
			}
		case ir.Property:
			if node.Name == "expression" {
				reads++
				if !program.CheckedRead(node) {
					t.Error("access.expression through the cast's result isn't a checked read")
				}
			}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if casts != 1 || reads == 0 {
		t.Fatalf("got %d checked casts and %d reads of expression, want 1 and some", casts, reads)
	}
}

// What the kind-union road leaves refused: in .a the cast is a check no proof erases, and its fix narrows by
// the tag; in .a a source whose tag can only be one of the target's kinds makes the test vacuous; a target
// holding an intersection is a later head; and a member that could be written, through the source, a slot it
// can't hold is no checked view.
func TestKindUnionCastStaysRefused(t *testing.T) {
	t.Parallel()
	cast := "function target(node: Node): Expression { return (node as PropertyAccessExpression | ElementAccessExpression).expression; }\n"
	for _, probe := range []struct{ name, source, extension, reason string }{
		{"in .a", kindUnionModel + cast, ".a", "node.kind !== SyntaxKind.PropertyAccessExpression"},
		{"in .a, both kinds", kindUnionModel + cast, ".a", "node.kind !== SyntaxKind.ElementAccessExpression"},
		{"vacuous in .a", `enum SyntaxKind { Identifier = 80, PropertyAccessExpression = 211, ElementAccessExpression = 212 }
interface Node { readonly kind: SyntaxKind.PropertyAccessExpression | SyntaxKind.ElementAccessExpression; readonly pos: number }
interface PropertyAccessExpression extends Node { readonly kind: SyntaxKind.PropertyAccessExpression; readonly expression: Node }
interface ElementAccessExpression extends Node { readonly kind: SyntaxKind.ElementAccessExpression; readonly expression: Node }
function target(node: Node): Node { return (node as PropertyAccessExpression | ElementAccessExpression).expression; }
`, ".a", "whose tag test is vacuous"},
		{"intersection member", kindUnionModel + "interface Extra { readonly extra: number }\nfunction target(node: Node): Expression { return (node as (PropertyAccessExpression & Extra) | ElementAccessExpression).expression; }\n", ".ts", "a cast the runtime can't check"},
		{"widened slot", `enum SyntaxKind { Identifier = 80, Leaf = 90, Other = 91 }
interface Node { readonly kind: SyntaxKind; parent: Node }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly text: string }
interface Leaf extends Node { readonly kind: SyntaxKind.Leaf; parent: Identifier }
interface Other extends Node { readonly kind: SyntaxKind.Other; readonly size: number }
function target(node: Node): Node { return (node as Leaf | Other).parent; }
`, ".ts", "a cast the runtime can't check"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want a refusal containing %q, got %v", probe.reason, err)
			}
		})
	}
}

// A write through the kind cast's union stays NotYet (kindUnionWrites): Adamic ruled it as #b5w3ycg's head 2,
// and this head keeps it waiting, written through the cast, through a local, through an alias of the union, and
// as a compound assignment.
func TestKindUnionCastWriteStaysNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, write string }{
		{"object field", "(node as PropertyAccessExpression | ElementAccessExpression).expression = id;\n"},
		{"scalar field", "(node as PropertyAccessExpression | ElementAccessExpression).flags = 1;\n"},
		{"through a local", "const access = node as PropertyAccessExpression | ElementAccessExpression;\naccess.expression = id;\n"},
		{"through an alias", "type AccessExpression = PropertyAccessExpression | ElementAccessExpression;\nconst access = node as AccessExpression;\naccess.flags += 1;\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, kindUnionWriteModel+probe.write, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "waits on #b5w3ycg head 2") {
				t.Fatalf("want NotYet at the write, waiting on head 2, got %v", err)
			}
		})
	}
}

// kindUnionWriteModel is kindUnionModel's shape with fields a write can reach.
const kindUnionWriteModel = `enum SyntaxKind { Identifier = 80, PropertyAccessExpression = 211, ElementAccessExpression = 212 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly text: string }
interface PropertyAccessExpression extends Node { readonly kind: SyntaxKind.PropertyAccessExpression; expression: Node; flags: number }
interface ElementAccessExpression extends Node { readonly kind: SyntaxKind.ElementAccessExpression; expression: Node; flags: number }
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 1, text: 'a' };
const built: PropertyAccessExpression = { kind: SyntaxKind.PropertyAccessExpression, pos: 0, expression: id, flags: 0 };
const node: Node = built;
`

// A write after narrowing the kind cast's result to one member goes on main's checked-write road through that
// member's view, the road main's one-kind cast already admits: it lowers, agrees with Node, and is a checked
// SetProperty, never a raw store. Head 1 adds no write road main doesn't have.
func TestKindUnionCastNarrowedWriteIsChecked(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, kindUnionWriteModel+`const access = node as PropertyAccessExpression | ElementAccessExpression;
if (access.kind === SyntaxKind.PropertyAccessExpression) {
	access.flags = 2;
}
console.log(String(access.flags) + ' ' + String(built.flags));
`, ".ts")
	writes := 0
	inspect := func(node any) bool {
		if write, ok := node.(ir.SetProperty); ok && write.Name == "flags" {
			writes++
			if !program.CheckedField(write.ViewReceiverTypeID, write.Name) {
				t.Error("the narrowed write of flags isn't a checked write")
			}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if writes != 1 {
		t.Fatalf("want one write of flags, found %d", writes)
	}
}

// tsc's recursive describe, a child read out of the cast's view passed back as a Node and switched on by kind, stays
// NotYet: the child's kind is then read through Node, which no view checks, and the child may be a value a view
// checked (#80cs7rz's widened read, checkViewWidenedReads). Lower's #zpzm9ta makes that read a checked one, and
// this pin goes when it lands. Until then the agreeing fixture, cast_kind_union.ts, prints a child through its view.
func TestKindUnionCastRecursiveReadStaysNotYet(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, kindUnionModel+`function describe(node: Node): string {
	switch (node.kind) {
		case SyntaxKind.PropertyAccessExpression:
		case SyntaxKind.ElementAccessExpression: {
			const access = node as PropertyAccessExpression | ElementAccessExpression;
			return describe(access.expression) + (access.kind === SyntaxKind.PropertyAccessExpression ? '.' + access.name.text : '[]');
		}
		case SyntaxKind.Identifier:
			return (node as Identifier).text;
	}
	return '?';
}
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 1, end: 2, text: 'a' };
const property: PropertyAccessExpression = { kind: SyntaxKind.PropertyAccessExpression, pos: 0, end: 3, expression: id, name: id };
console.log(describe(property));
`, ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "which may hold a checked view's value") {
		t.Fatalf("want NotYet at the read through Node, which may hold a checked view's value, got %v", err)
	}
}

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

// tsc's recursive describe, a child read out of the cast's view passed back as a Node and switched on by kind: the
// child's kind is then read through Node, which no view checks, and the child may be a value a view checked
// (#80cs7rz's widened read, checkViewWidenedReads). It was NotYet there until #zpzm9ta made that read checked where
// it's made, so it lowers now and agrees with Node, natively too. Mutant: readiness stripping a marked read's view
// anyway, and widenedReadsOf fails it.
func TestKindUnionCastRecursiveReadIsChecked(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, kindUnionModel+`function describe(node: Node): string {
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
	if checked, _ := widenedReadsOf(t, program, "kind"); checked == 0 {
		t.Fatal("describe's node.kind isn't a widened read checked where it's made")
	}
	if code, out, errs := brandedNative(t, program); code != 0 || out != "a.a\n" {
		t.Fatalf("sanitized native exit %d, stdout %q, stderr %.300q; want a.a, as Node prints", code, out, errs)
	}
}

// Lower's read of head 1 asked which road holds each way a value cast to the kind union can be written or tested
// other than a plain write through the union, and the getter-backed tag. Each is measured where it stops on main
// (J-y4a6efa and J-gzyv3tp, the probes; J-9caypx8 for the shared-kind lie's narrowed write, which main's checked
// write stops at the write with exit 70 on both backends, and which #b5w3ycg head 2 pins as an agreement witness).
// A road that starts lowering turns its row red first, so whoever opens it reads this. The increment's row went
// when head 2 opened it: a named write, increments included, lowers and agrees with Node there
// (cast_kind_union_writes_test.go's TestKindUnionWriteLowers).
func TestKindUnionCastRoadsLowerAsked(t *testing.T) {
	t.Parallel()
	cast := "const access = node as PropertyAccessExpression | ElementAccessExpression;\n"
	for _, probe := range []struct{ name, source, words string }{
		{"Object.assign of a scalar", kindUnionWriteModel + cast + "Object.assign(access, { flags: 5 });\nconsole.log(String(access.flags));\n", "Object.assign on a shape not proven"},
		{"Object.assign of an object", kindUnionWriteModel + cast + "Object.assign(access, { expression: id });\nconsole.log(String(access.expression.kind));\n", "Object.assign on a shape not proven"},
		{"destructuring assignment target", kindUnionWriteModel + cast + "const src: { a: Node } = { a: id };\n({ a: access.expression } = src);\nconsole.log(String(access.expression.kind));\n", "which #b5w3ycg head 2 leaves waiting"},
		{"for-of target", kindUnionWriteModel + cast + "const values: number[] = [1, 2];\nfor (access.flags of values) { console.log('x'); }\nconsole.log(String(access.flags));\n", "which #b5w3ycg head 2 leaves waiting"},
		{"a method on the shared-kind lie", `enum SyntaxKind { Identifier = 80, PrivateIdentifier = 81, PropertyDeclaration = 172, MethodDeclaration = 174 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly escapedText: string }
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 2, escapedText: 'visit' };
interface MethodDeclaration extends Node { readonly kind: SyntaxKind.MethodDeclaration; readonly name: Identifier; count: number }
interface TaggedMethodDeclaration extends MethodDeclaration { tag: number; retag(): void }
interface PropertyDeclaration extends Node { readonly kind: SyntaxKind.PropertyDeclaration; readonly name: Identifier; tag: number; retag(): void }
function mark(node: Node): void {
	const member = node as TaggedMethodDeclaration | PropertyDeclaration;
	if (member.kind === SyntaxKind.MethodDeclaration) {
		member.retag();
	}
}
const method: MethodDeclaration = { kind: SyntaxKind.MethodDeclaration, pos: 1, name: id, count: 0 };
console.log('marking');
mark(method);
console.log(String(method.count));
`, "stage 0 can't lower a callable field retag read through a checked view, which needs a check at the call (rulings.md:27) yet"},
		{"a getter-backed tag", kindUnionModel + `function target(node: Node): Expression { return (node as PropertyAccessExpression | ElementAccessExpression).expression; }
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 1, end: 2, text: 'a' };
const property: PropertyAccessExpression = { get kind(): SyntaxKind.PropertyAccessExpression { return SyntaxKind.PropertyAccessExpression; }, pos: 0, end: 3, expression: id, name: id };
console.log(String(target(property).pos));
`, "a getter in a checked field contract"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			if err == nil || !strings.Contains(err.Error(), probe.words) {
				t.Fatalf("want %q, got %v", probe.words, err)
			}
		})
	}
}

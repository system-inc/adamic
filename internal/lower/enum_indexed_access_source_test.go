package lower

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// An indexed access over a type parameter, T["kind"] with T extends Node, is read at its base constraint before
// widened judges it, as a type parameter source is (#snypn8c c5', Compiler's application of Adamic's c5 ruling,
// enum_type_parameter_source_test.go). tsc's parser.ts createMissingNode<T extends Node>(kind: T["kind"]) hands
// kind to createToken<TKind extends SyntaxKind>, then relates finishNode's Token<T["kind"]> to Node: the open enum
// arm asked enumAssignable of T["kind"] itself, which isn't a number, and refused both as an unproven value. The
// read fails closed for an access over anything but a bare type parameter and for a constraint that is still
// generic, and a mapped source is no indexed access, so nodeFactory's update over Mutable<T> stays refused. A value
// of type T["kind"] has no lowering yet, a NotYet of its own, so the two shapes the read admits are held to that
// NotYet, which lowering reaches only once the refusal pass lets them through. Mutants
// (enum_indexed_access_source.mutants.md): the read removed; the object's type parameter guard dropped; the
// generic constraint guard dropped.

// accessTokenSource is createMissingNode's token, small: SyntaxNode is tsc's Node, token is createToken's last
// overload, and make hands its T["kind"] to it, whose binding constraintRelation judges against E.
const accessTokenSource = `enum E { A, B, C }
interface SyntaxNode {
	readonly kind: E;
}
interface Leaf extends SyntaxNode {
	readonly kind: E.B;
}
function token<K extends E>(kind: K): K {
	return kind;
}
function make<T extends SyntaxNode>(kind: T["kind"]) {
	return token(kind);
}
console.log(String(make<Leaf>(E.B)));
console.log(String(make<SyntaxNode>(E.C)));
`

// accessUnionSource is createMissingNode's finishNode: the result is a union holding Token<T["kind"]>, bound to
// finish's T extends SyntaxNode, so each member is related to SyntaxNode, and Token's kind to E.
const accessUnionSource = `enum E { A, B, C }
interface SyntaxNode {
	readonly kind: E;
}
interface Named extends SyntaxNode {
	readonly kind: E.A;
	readonly text: string;
}
interface Token<K extends E> extends SyntaxNode {
	readonly kind: K;
}
function finish<T extends SyntaxNode>(node: T): T {
	return node;
}
function makeNamed(): Named {
	return { kind: E.A, text: 'n' };
}
function token<K extends E>(kind: K): Token<K> {
	const made: Token<K> = { kind };
	return made;
}
function missing<T extends SyntaxNode>(kind: T["kind"], named: boolean): SyntaxNode {
	const result = named ? makeNamed() : token(kind);
	return finish(result);
}
console.log(String(missing<Named>(E.A, true).kind));
console.log(String(missing<SyntaxNode>(E.C, false).kind));
`

// refusedIntoEnumSlot reports whether err is the refusal of an unproven value written into a numeric enum slot, at
// where, whose slot is printed as exactly members, in any order, since the checker prints a union's members by
// their ids.
func refusedIntoEnumSlot(err error, where string, members ...string) bool {
	var refused *Refused
	if !errors.As(err, &refused) {
		return false
	}
	// Where may name the file by its whole path, so the message is read from where's own name on.
	path, slot, found := strings.Cut(err.Error(), where+": Adamic 0.1 refuses an unproven value assigned to a numeric literal or enum member slot ")
	if !found || path != "" && !strings.HasSuffix(path, "/") {
		return false
	}
	if slot, found = strings.CutSuffix(slot, "; compare with this literal and return that constant, or widen the slot to the whole numeric enum or number (adamic/enum-literal)"); !found {
		return false
	}
	printed, want := strings.Split(slot, " | "), slices.Clone(members)
	slices.Sort(printed)
	slices.Sort(want)
	return slices.Equal(printed, want)
}

// withAccessChecker loads source and hands judge a lowering over its checker, with the file's statements.
func withAccessChecker(t *testing.T, source string, judge func(l *lowering, statements []*ast.Node)) {
	t.Helper()
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	typeChecker, release := program.Checker(context.Background(), file)
	defer release()
	judge(&lowering{checker: typeChecker, program: program}, file.Statements.Nodes)
}

// namedStatement is the top-level statement of a kind declaring name, or nil.
func namedStatement(statements []*ast.Node, kind ast.Kind, name string) *ast.Node {
	for _, statement := range statements {
		if statement.Kind == kind && statement.Name() != nil && statement.Name().Text() == name {
			return statement
		}
	}
	return nil
}

// passesTheRefusalPass fails t unless lowering source stops at the NotYet want ends with, which lowering meets only
// after the refusal pass, and never at the unproven value refusal the read lifts.
func passesTheRefusalPass(t *testing.T, source, want string) {
	t.Helper()
	_, err := lowerSource(t, source)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), want) || strings.Contains(err.Error(), "enum member slot") {
		t.Fatalf("got %v, want the not yet ending %q", err, want)
	}
}

// The token shape passes the refusal pass: make's T["kind"] bound to token's K extends E is read at E, so make stops
// only where lowering meets a function returning T["kind"], which it can't lower yet.
func TestIndexedAccessSourceIntoWholeEnumSlotLowers(t *testing.T) {
	t.Parallel()
	passesTheRefusalPass(t, accessTokenSource, "main.a:11:10: stage 0 can't lower a function returning T[\"kind\"] yet")
}

// The relation the token shape reaches, asked of widened itself: T["kind"] seen as E passes, and the open enum
// arm alone, enumAssignable of T["kind"], still says no, so the pass is the read and nothing else.
func TestIndexedAccessSourceIsReadAtItsConstraint(t *testing.T) {
	t.Parallel()
	withAccessChecker(t, accessTokenSource, func(l *lowering, statements []*ast.Node) {
		enum, maker := namedStatement(statements, ast.KindEnumDeclaration, "E"), namedStatement(statements, ast.KindFunctionDeclaration, "make")
		if enum == nil || maker == nil {
			t.Fatalf("no E (%v) or make (%v) in the source", enum, maker)
		}
		open, kind := l.checker.GetTypeAtLocation(enum.Name()), l.checker.GetTypeAtLocation(maker.Parameters()[0].Name())
		if kind.Flags()&checker.TypeFlagsIndexedAccess == 0 || !typeParameterAccess(kind) || !l.openNumericEnumType(open) {
			t.Fatalf("make's kind is %s and E %s, want an indexed access over a type parameter and an open numeric enum", l.checker.TypeToString(kind), l.checker.TypeToString(open))
		}
		if l.enumAssignable(kind, open) {
			t.Fatalf("enumAssignable(T[\"kind\"], E) holds, so this pin no longer proves the read")
		}
		if found := l.widened(kind, open, map[[2]*checker.Type]bool{}); found != nil {
			t.Fatalf("T[\"kind\"] seen as E: widened found %s seen as %s, want nothing", l.checker.TypeToString(found.source), l.checker.TypeToString(found.target))
		}
	})
}

// The finishNode shape passes the refusal pass: the union holding Token<T["kind"]> is related to SyntaxNode member
// by member, and Token's kind is read at E, so missing stops only where lowering meets its parameter of type
// T["kind"], which it can't lower yet.
func TestIndexedAccessSourceInUnionMemberLowers(t *testing.T) {
	t.Parallel()
	passesTheRefusalPass(t, accessUnionSource, "main.a:22:40: stage 0 can't lower a value of type T[\"kind\"], an unresolved indexed access yet")
}

// The union's own road, asked of widened: the result's union holds a member whose kind is an indexed access over
// a type parameter, and the whole union seen as SyntaxNode passes. The program above also hands T["kind"] to
// token, the first pin's road, so this holds the union's relation apart from it.
func TestIndexedAccessSourceInUnionMemberIsReadAtItsConstraint(t *testing.T) {
	t.Parallel()
	withAccessChecker(t, accessUnionSource, func(l *lowering, statements []*ast.Node) {
		missing, node := namedStatement(statements, ast.KindFunctionDeclaration, "missing"), namedStatement(statements, ast.KindInterfaceDeclaration, "SyntaxNode")
		if missing == nil || node == nil {
			t.Fatalf("no missing (%v) or SyntaxNode (%v) in the source", missing, node)
		}
		result := missing.Body().AsBlock().Statements.Nodes[0].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Name()
		union, slot := l.checker.GetTypeAtLocation(result), l.checker.GetDeclaredTypeOfSymbol(l.checker.GetSymbolOfDeclaration(node))
		accessed := false
		for _, member := range l.definedMembers(union) {
			if property := l.checker.GetPropertyOfType(member, "kind"); property != nil {
				if kind := l.checker.GetTypeOfSymbol(property); kind.Flags()&checker.TypeFlagsIndexedAccess != 0 && typeParameterAccess(kind) {
					accessed = true
				}
			}
		}
		if len(l.definedMembers(union)) < 2 || !accessed {
			t.Fatalf("result is %s, want a union with a member whose kind is T[\"kind\"]", l.checker.TypeToString(union))
		}
		if found := l.widened(union, slot, map[[2]*checker.Type]bool{}); found != nil {
			t.Fatalf("%s seen as SyntaxNode: widened found %s seen as %s, want nothing", l.checker.TypeToString(union), l.checker.TypeToString(found.source), l.checker.TypeToString(found.target))
		}
	})
}

// A read is never a pass of its own (docs/rulings.md:247): an access whose constraint is a number, into a member
// slot, is a number there and refused in exactly the bare number's words (memberSlotRefusal, c5's pin). A
// constraint that names a member other than the slot's, T extends { kind: E.A } into E.B, is one tsc itself
// rejects, so the narrower slot is pinned with the constraint tsc lets in.
func TestIndexedAccessSourceIntoMemberSlotIsRefusedAsItsConstraint(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, signature string }{
		{"constrained to a number", "function slot<T extends { kind: number }>(value: T[\"kind\"]): E.A {"},
		{"constrained to a number beside a member", "function slot<T extends { kind: number | E.A }>(value: T[\"kind\"]): E.A {"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := "enum E { A, B, C }\n" + probe.signature + "\n\treturn value;\n}\n"
			_, err := lowerSource(t, source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), memberSlotRefusal) {
				t.Fatalf("got %v, want a refusal ending %q", err, memberSlotRefusal)
			}
		})
	}
}

// nodeFactory's finishUpdateBaseSignatureDeclaration<T>(updated: Mutable<T>, original: T) stays refused where it
// hands its Mutable<T> to update<T extends Node>(updated: T, original: T), as stage3's
// 99-memory-factory-update-view leaves update. The walk of Mutable<T> seen as SyntaxNode reaches T["kind"] as the
// mapped type's template, which the read now admits, and beside it the writable original, and the argument is
// Mutable<T> seen as T, a type parameter whose constraint can be written, which another ruling owes. The pin holds
// that the shape stays refused, whichever of those stops it first, and logs the words it stops with.
func TestIndexedAccessSourceThroughMappedTypeStaysRefused(t *testing.T) {
	t.Parallel()
	source := `enum E { A, B, C }
interface SyntaxNode {
	readonly kind: E;
	original?: SyntaxNode;
}
type Mutable<T> = { -readonly [K in keyof T]: T[K] };
function update<T extends SyntaxNode>(updated: T, original: T): T {
	return updated === original ? original : updated;
}
function finishUpdate<T extends SyntaxNode>(updated: Mutable<T>, original: T): T {
	return update(updated, original);
}
`
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) {
		t.Fatalf("got %v, want a refusal", err)
	}
	t.Logf("finishUpdate stops with: %v", err)
}

// A constraint that is itself generic fails closed: T extends { kind: U } gives T["kind"] the constraint U, so
// it isn't read, and token's binding is refused as on main, though U extends E would read at E (Compiler's
// condition on #snypn8c c5').
func TestIndexedAccessSourceWithGenericConstraintStaysRefused(t *testing.T) {
	t.Parallel()
	source := "enum E { A, B, C }\nfunction token<K extends E>(kind: K): K {\n\treturn kind;\n}\nfunction make<U extends E, T extends { kind: U }>(kind: T[\"kind\"]) {\n\treturn token(kind);\n}\n"
	if _, err := lowerSource(t, source); !refusedIntoEnumSlot(err, "main.a:6:15", "E") {
		t.Fatalf("got %v, want the unproven value refusal at main.a:6:15 into the slot E", err)
	}
}

// An indexed access over an object that isn't a type parameter fails closed: SyntaxNode[P] with P extends "kind"
// stays an indexed access to the checker, since P is generic, and its base constraint is E, but it isn't read,
// and token's binding is refused as on main (Compiler's condition on #snypn8c c5').
func TestIndexedAccessSourceOverAnObjectStaysRefused(t *testing.T) {
	t.Parallel()
	source := "enum E { A, B, C }\ninterface SyntaxNode {\n\treadonly kind: E;\n}\nfunction token<K extends E>(kind: K): K {\n\treturn kind;\n}\nfunction make<P extends \"kind\">(kind: SyntaxNode[P]) {\n\treturn token(kind);\n}\n"
	if _, err := lowerSource(t, source); !refusedIntoEnumSlot(err, "main.a:9:15", "E") {
		t.Fatalf("got %v, want the unproven value refusal at main.a:9:15 into the slot E", err)
	}
}

// The two lies tsc writes into a member slot, each a value the types don't prove, stay refused (Adamic's ruling on
// #snypn8c), lowered as .ts as tsc is:
//   - builder.ts's getNextAffectedFilePendingEmit: a BuilderFileEmit narrowed by truthiness, written where the
//     enum without its zero member goes. A numeric enum holds any number, so a truthy one may be no member.
//   - parenthesizerRules.ts's getLiteralKindOfBinaryPlusOperand: a SyntaxKind narrowed by isLiteralKind, whose body
//     is a range test, then a conditional's branch, whose type is a narrower union of members.
func TestIndexedAccessHeadLiesStayRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source, where string
		slot                []string
	}{
		{
			"a truthy kind into the enum without its zero member",
			"enum Emit { None, Js, Dts }\nfunction pending(): Emit {\n\treturn Emit.Js;\n}\nfunction next(): { emitKind: Emit.Js | Emit.Dts } | undefined {\n\tconst pendingKind: Emit = pending();\n\tif (pendingKind) return { emitKind: pendingKind };\n\treturn undefined;\n}\n",
			"main.ts:7:38",
			[]string{"Emit.Js", "Emit.Dts"},
		},
		{
			"a predicate's kind into a narrower union of members",
			"enum Kind { Unknown, Numeric, Text, Plus }\nfunction isLiteralKind(kind: Kind): kind is Kind.Numeric | Kind.Text {\n\treturn Kind.Numeric <= kind && kind <= Kind.Text;\n}\nfunction kindOf(text: string): Kind {\n\treturn text === '1' ? Kind.Numeric : Kind.Plus;\n}\nfunction literalKind(left: string, right: string): Kind {\n\tconst leftKind = kindOf(left);\n\tconst literal = isLiteralKind(leftKind) && leftKind === kindOf(right) ? leftKind : Kind.Unknown;\n\treturn literal;\n}\n",
			"main.ts:10:74",
			[]string{"Kind.Unknown", "Kind.Numeric", "Kind.Text"},
		},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, probe.source, ".ts"); !refusedIntoEnumSlot(err, probe.where, probe.slot...) {
				t.Fatalf("got %v, want the unproven value refusal at %s into the slot %s", err, probe.where, strings.Join(probe.slot, " | "))
			}
		})
	}
}

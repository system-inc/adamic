package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// A source whose type is a type parameter is read at its base constraint before widened judges it, and
// then judged by every arm as a value of that constraint's type is (#snypn8c c5, Adamic's ruling). tsc's
// createToken (nodeFactory.ts:1458) writes createBaseToken<Token<TKind>>(token), and the written binding
// Token<TKind> against T extends Node relates Token's kind, TKind extends SyntaxKind, to Node's kind, the
// whole open enum: the open enum arm asked enumAssignable of TKind itself, which isn't a number, and refused
// it as an unproven value. Mutants (c5-mutants.md): the constraint read skipped, under which the token shape
// is refused again; and a type parameter source passed whole, under which T extends number and a chain to
// number go into a member slot unrefused.

// tokenSource is nodeFactory.ts:1458's shape, small: SyntaxNode and Token are tsc's Node and Token, and
// base is createBaseToken, whose written type argument constraintRelation judges against SyntaxNode.
const tokenSource = `enum SyntaxKind { Unknown, Identifier, Keyword }
interface SyntaxNode {
	readonly kind: SyntaxKind;
}
interface Token<TKind extends SyntaxKind> extends SyntaxNode {
	readonly kind: TKind;
}
function base<T extends SyntaxNode>(node: T): T {
	return node;
}
function make<TKind extends SyntaxKind>(kind: TKind): Token<TKind> {
	const token: Token<TKind> = { kind };
	return base<Token<TKind>>(token);
}
console.log(String(make(SyntaxKind.Keyword).kind));
console.log(String(make(SyntaxKind.Identifier).kind));
`

// The token shape lowers and agrees with Node: Token<TKind> bound to T extends SyntaxNode relates TKind to
// the whole SyntaxKind, which its constraint is.
func TestTypeParameterSourceIntoWholeEnumSlotLowers(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, tokenSource)
}

// The relation the token shape reaches, asked of widened itself: TKind seen as SyntaxKind passes, and the
// open enum arm alone, enumAssignable of TKind, still says no, so the pass is the constraint read and
// nothing else. That's how the shape above is known to reach this relation, not some other road.
func TestTypeParameterSourceIsReadAtItsConstraint(t *testing.T) {
	t.Parallel()
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", tokenSource)})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	typeChecker, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{checker: typeChecker, program: program}
	var kind, open *checker.Type
	for _, statement := range file.Statements.Nodes {
		switch {
		case statement.Kind == ast.KindEnumDeclaration && statement.Name().Text() == "SyntaxKind":
			open = typeChecker.GetTypeAtLocation(statement.Name())
		case statement.Kind == ast.KindFunctionDeclaration && statement.Name().Text() == "make":
			kind = typeChecker.GetTypeAtLocation(statement.TypeParameters()[0].Name())
		}
	}
	if kind == nil || open == nil {
		t.Fatalf("no TKind (%v) or SyntaxKind (%v) in the source", kind, open)
	}
	if kind.Flags()&checker.TypeFlagsTypeParameter == 0 || !l.openNumericEnumType(open) {
		t.Fatalf("TKind is %s and SyntaxKind %s, want a type parameter and an open numeric enum", typeChecker.TypeToString(kind), typeChecker.TypeToString(open))
	}
	if l.enumAssignable(kind, open) {
		t.Fatalf("enumAssignable(TKind, SyntaxKind) holds, so this pin no longer proves the constraint read")
	}
	if found := l.widened(kind, open, map[[2]*checker.Type]bool{}); found != nil {
		t.Fatalf("TKind seen as SyntaxKind: widened found %s seen as %s, want nothing", typeChecker.TypeToString(found.source), typeChecker.TypeToString(found.target))
	}
}

// memberSlotRefusal is what a plain number returned into the member slot E.A is refused with, at line 3,
// column 9 of each source below, all laid out alike.
const memberSlotRefusal = "main.a:3:9: Adamic 0.1 refuses an unproven value assigned to a numeric literal or enum member slot E.A; compare with this literal and return that constant, or widen the slot to the whole numeric enum or number (adamic/enum-literal)"

// A type parameter is never a pass of its own (docs/rulings.md:247): T extends number into a member slot is a
// number there, refused in exactly the bare number's words, and so is a chain to number, U extends T with
// T extends number, read to its base.
func TestTypeParameterSourceIntoMemberSlotIsRefusedAsItsConstraint(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, signature string }{
		{"number", "function slot(value: number): E.A {"},
		{"constrained to number", "function slot<T extends number>(value: T): E.A {"},
		{"a chain to number", "function slot<T extends number, U extends T>(value: U): E.A {"},
		// Language's three on #snypn8c, none through a chain of type parameters, so a later reorder of
		// widened's arms meets these first.
		{"constrained to a union of number and a member", "function slot<T extends number | E.A>(value: T): E.A {"},
		{"a conditional type", "function slot<T>(value: T extends number ? T : never): E.A {"},
		{"constrained to an intersection", "function slot<T extends number & { readonly brand?: never }>(value: T): E.A {"},
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

// T extends E.A is E.A, so into an E.A slot it lowers and agrees with Node.
func TestTypeParameterSourceConstrainedToTheMemberLowers(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, "enum E { A, B, C }\nfunction slot<T extends E.A>(value: T): E.A {\n\treturn value;\n}\nconsole.log(String(slot(E.A)));\n")
}

// A type parameter with no constraint has none to be read at, and is judged as on main: widened passes it,
// here into an unknown slot.
func TestUnconstrainedTypeParameterSourceIsJudgedAsBefore(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, "function describe<T>(value: T): string {\n\tconst seen: unknown = value;\n\treturn typeof seen;\n}\nconsole.log(describe(7));\nconsole.log(describe('seven'));\n")
}

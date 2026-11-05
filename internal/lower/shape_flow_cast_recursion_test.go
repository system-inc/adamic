package lower

import (
	"testing"
	"time"
)

// binder.ts's isNarrowableOperand recurses on a field of its own parameter seen through a cast,
// `isNarrowableOperand((expr as BinaryExpression).left)`, so the parameter's sources are property
// reads whose receiver is a CheckedCast of the parameter, never a bare Read, and only a bare Read's
// projected query was memoized. Every view origin's query reopened a frame for the cast receiver at
// every level, once per recursive call: three calls make 3^32 frames before the depth cap, and
// lowering never finished (#tzy22kd, found by Performance's unit-C instrument on binder.ts; main
// stopped at binder.ts:1341). Each subtype narrows kind to a literal member, as tsc's do, so each cast
// is a tag-checked CheckedCast of the parameter, which is the receiver binder's query stack repeats
// (a pin whose subtypes left kind wide never built that receiver, and passed with the memos dropped).
// Lowering must finish well inside the budget, and the program agree with Node, natively and on
// JavaScript. Mutant: the frame memo and the follow dedupe both dropped, so the budget runs out.
func TestShapeFlowCastRecursionFinishes(t *testing.T) {
	t.Parallel()
	const source = `enum SyntaxKind { Identifier, ParenthesizedExpression, NonNullExpression, BinaryExpression, EqualsToken, CommaToken }
interface Node { readonly kind: SyntaxKind; }
interface Expression extends Node { readonly _expressionBrand: undefined; }
interface Token extends Node { readonly kind: SyntaxKind.EqualsToken | SyntaxKind.CommaToken; }
interface Identifier extends Expression { readonly kind: SyntaxKind.Identifier; }
interface ParenthesizedExpression extends Expression { readonly kind: SyntaxKind.ParenthesizedExpression; readonly expression: Expression; }
interface NonNullExpression extends Expression { readonly kind: SyntaxKind.NonNullExpression; readonly expression: Expression; }
interface BinaryExpression extends Expression { readonly kind: SyntaxKind.BinaryExpression; readonly left: Expression; readonly operatorToken: Token; readonly right: Expression; }
function narrowable(expr: Expression): boolean {
	switch (expr.kind) {
		case SyntaxKind.ParenthesizedExpression:
			return narrowable((expr as ParenthesizedExpression).expression);
		case SyntaxKind.NonNullExpression:
			return narrowable((expr as NonNullExpression).expression);
		case SyntaxKind.BinaryExpression:
			switch ((expr as BinaryExpression).operatorToken.kind) {
				case SyntaxKind.EqualsToken:
					return narrowable((expr as BinaryExpression).left);
				case SyntaxKind.CommaToken:
					return narrowable((expr as BinaryExpression).right);
			}
	}
	return expr.kind === SyntaxKind.Identifier;
}
const id: Identifier = { kind: SyntaxKind.Identifier, _expressionBrand: undefined };
const comma: Token = { kind: SyntaxKind.CommaToken };
const paren: ParenthesizedExpression = { kind: SyntaxKind.ParenthesizedExpression, _expressionBrand: undefined, expression: id };
const binary: BinaryExpression = { kind: SyntaxKind.BinaryExpression, _expressionBrand: undefined, left: id, operatorToken: comma, right: paren };
console.log(String(narrowable(binary)));
const nonNull: NonNullExpression = { kind: SyntaxKind.NonNullExpression, _expressionBrand: undefined, expression: binary };
console.log(String(narrowable(nonNull)));
`
	done := make(chan error, 1)
	go func() {
		_, err := lowerSourceAs(t, source, ".ts")
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("want it lowered, got %v", err)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("lowering the cast recursion took longer than 20 seconds")
	}
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "true\ntrue\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "true\ntrue\n")
	}
}

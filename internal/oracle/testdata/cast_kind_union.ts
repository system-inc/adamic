// tsc's tag interface downcast to a union of kinds (#b5w3ycg slice 4, Adamic's Nov 3 ruling): a Node whose kind is
// the open SyntaxKind, switched on and cast to the kinded interfaces its case names, as checker.ts casts
// `node as PropertyAccessExpression | ElementAccessExpression`. The cast tests the kind against exactly the target's
// kinds, and every read through the result is checked by the members' views: a common field, an object field, a
// field of one member after narrowing on the kind, from a Node, from an Expression, and to a union alias. Every
// value is built as its own kind, so every cast holds and the program prints what Node prints. A child read out of
// a view is printed through that view, never passed back as a Node: a read of it through Node, as tsc's recursive
// describe makes, waits on #zpzm9ta (#80cs7rz's widened read, pinned in cast_kind_union_test.go).
enum SyntaxKind { NumericLiteral = 9, Identifier = 80, PropertyAccessExpression = 211, ElementAccessExpression = 212, CallExpression = 213 }
interface Node { readonly kind: SyntaxKind; readonly pos: number; readonly end: number }
interface Expression extends Node { readonly parenthesized: boolean }
interface Identifier extends Expression { readonly kind: SyntaxKind.Identifier; readonly escapedText: string }
interface NumericLiteral extends Expression { readonly kind: SyntaxKind.NumericLiteral; readonly text: string }
interface PropertyAccessExpression extends Expression { readonly kind: SyntaxKind.PropertyAccessExpression; readonly expression: Expression; readonly name: Identifier }
interface ElementAccessExpression extends Expression { readonly kind: SyntaxKind.ElementAccessExpression; readonly expression: Expression; readonly argumentExpression: Expression }
interface CallExpression extends Expression { readonly kind: SyntaxKind.CallExpression; readonly expression: Expression; readonly argumentCount: number }
type AccessExpression = PropertyAccessExpression | ElementAccessExpression;

function describe(node: Node): string {
	switch (node.kind) {
		case SyntaxKind.PropertyAccessExpression:
		case SyntaxKind.ElementAccessExpression: {
			const access = node as PropertyAccessExpression | ElementAccessExpression;
			const tail = access.kind === SyntaxKind.PropertyAccessExpression ? '.' + access.name.escapedText : '[' + String(access.argumentExpression.kind) + ']';
			return String(access.expression.kind) + '@' + String(access.expression.pos) + '..' + String(access.expression.end) + tail;
		}
		case SyntaxKind.Identifier:
			return (node as Identifier).escapedText;
		case SyntaxKind.NumericLiteral:
			return (node as NumericLiteral).text;
	}
	return '?';
}
function span(node: Node): string {
	const access = node as AccessExpression;
	return String(access.pos) + '..' + String(access.end);
}
function callee(expression: Expression): string {
	const called = expression as PropertyAccessExpression | CallExpression;
	return String(called.expression.kind) + ' ' + String(called.expression.parenthesized) + ' ' + String(called.parenthesized);
}

const ts: Identifier = { kind: SyntaxKind.Identifier, pos: 0, end: 2, parenthesized: false, escapedText: 'ts' };
const factory: Identifier = { kind: SyntaxKind.Identifier, pos: 3, end: 10, parenthesized: false, escapedText: 'factory' };
const zero: NumericLiteral = { kind: SyntaxKind.NumericLiteral, pos: 11, end: 12, parenthesized: false, text: '0' };
const property: PropertyAccessExpression = { kind: SyntaxKind.PropertyAccessExpression, pos: 0, end: 10, parenthesized: false, expression: ts, name: factory };
const element: ElementAccessExpression = { kind: SyntaxKind.ElementAccessExpression, pos: 0, end: 13, parenthesized: true, expression: property, argumentExpression: zero };
const call: CallExpression = { kind: SyntaxKind.CallExpression, pos: 0, end: 15, parenthesized: false, expression: element, argumentCount: 0 };
console.log(describe(property));
console.log(describe(element));
console.log(span(property) + ' ' + span(element));
console.log(callee(property) + ' ' + callee(call));
console.log(describe(ts) + ' ' + describe(zero) + ' ' + describe(call));

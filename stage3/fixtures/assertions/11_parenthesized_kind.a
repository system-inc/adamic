// From TypeScript 6.0.3, src/compiler/utilities.ts:12056
// Census reason: type assertion sites
// Support types and driver are reduced scaffolding; the upstream function is unchanged.

const SyntaxKind = { ObjectLiteralExpression: 211, ArrayLiteralExpression: 210, ParenthesizedExpression: 218, ThisKeyword: 110, SuperKeyword: 108, PropertyAccessExpression: 212, ElementAccessExpression: 213 } as const;
type SyntaxKind = number;
interface Node { readonly kind: SyntaxKind; }
interface ObjectLiteralExpression extends Node { readonly kind: 211; readonly properties: readonly number[]; }
interface ArrayLiteralExpression extends Node { readonly kind: 210; readonly elements: readonly number[]; }
interface ParenthesizedExpression extends Node { readonly kind: 218; readonly expression: Expression; }
type Expression = Node;

export function unwrapParenthesizedExpression(o: Expression): Expression {
    while (o.kind === SyntaxKind.ParenthesizedExpression) {
        o = (o as ParenthesizedExpression).expression;
    }
    return o;
}

const leaf: Node = { kind: 110 };
const inner: ParenthesizedExpression = { kind: 218, expression: leaf };
const outer: ParenthesizedExpression = { kind: 218, expression: inner };
console.log(String(unwrapParenthesizedExpression(outer).kind));
console.log(String(unwrapParenthesizedExpression(leaf).kind));

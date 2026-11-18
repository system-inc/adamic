// The lie: a Node whose kind is outside the target's kinds, cast to the union of kinds. The cast's kind test fails
// with the cast's own message. Node runs on without a check and prints object.
enum SyntaxKind { Identifier = 80, PropertyAccessExpression = 211, ElementAccessExpression = 212 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly text: string }
interface PropertyAccessExpression extends Node { readonly kind: SyntaxKind.PropertyAccessExpression; readonly expression: Node }
interface ElementAccessExpression extends Node { readonly kind: SyntaxKind.ElementAccessExpression; readonly expression: Node }
function cast(node: Node): PropertyAccessExpression | ElementAccessExpression { return node as PropertyAccessExpression | ElementAccessExpression; }
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 0, text: 'a' };
const access: PropertyAccessExpression = { kind: SyntaxKind.PropertyAccessExpression, pos: 1, expression: id };
console.log('casting');
console.log(typeof cast(id) + ' ' + String(access.pos));

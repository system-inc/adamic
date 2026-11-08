// tsc's getAssignedTypeOfSpreadExpression (checker.ts:28349): a readonly union parent cast to one member.
interface Node { readonly kind: number; readonly pos: number; }
interface Expression extends Node { readonly brandExpression: true; }
interface ArrayLiteralExpression extends Expression { readonly kind: 210; readonly elements: readonly Expression[]; }
interface CallExpression extends Expression { readonly kind: 214; readonly callee: Expression; }
interface NewExpression extends Expression { readonly kind: 215; readonly callee: Expression; }
interface SpreadElement extends Expression { readonly kind: 231; readonly parent: ArrayLiteralExpression | CallExpression | NewExpression; }
function assigned(node: ArrayLiteralExpression): number {
	return node.elements.length;
}
function spread(node: SpreadElement): number {
	return assigned(node.parent as ArrayLiteralExpression);
}
const array: ArrayLiteralExpression = { kind: 210, pos: 0, brandExpression: true, elements: [] };
const element: SpreadElement = { kind: 231, pos: 1, brandExpression: true, parent: array };
console.log(spread(element));
export {};

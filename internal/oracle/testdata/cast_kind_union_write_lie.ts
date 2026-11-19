// The lie (#b5w3ycg head 2): a value built as a Node, whose kind is the open SyntaxKind, so no declared type the
// write's judgment reads carries it, holding kind ElementAccessExpression and no flags. Cast to the union and
// written, it is refused at compile time, since the value reaching the write holds no flags. Node runs on.
enum SyntaxKind { Identifier = 80, PropertyAccessExpression = 211, ElementAccessExpression = 212 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface PropertyAccessExpression extends Node { readonly kind: SyntaxKind.PropertyAccessExpression; flags: number }
interface ElementAccessExpression extends Node { readonly kind: SyntaxKind.ElementAccessExpression; flags: number }
function mark(node: Node): void {
	(node as PropertyAccessExpression | ElementAccessExpression).flags = 5;
}
const lie: Node = { kind: SyntaxKind.ElementAccessExpression, pos: 0 };
console.log('writing');
mark(lie);
console.log('after');

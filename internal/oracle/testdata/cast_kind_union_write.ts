// A named write through a kind-checked cast's union (#b5w3ycg head 2, Adamic's ruling of Nov 17): every declared
// type carrying one of the cast's kinds declares flags as a number, so the write goes on main's checked-write road.
// Then a value cast to the union, passed back as a Node and written through Node's own writable pos (Compiler's
// second pin). Both agree with Node.
enum SyntaxKind { Identifier = 80, PropertyAccessExpression = 211, ElementAccessExpression = 212 }
interface Node { readonly kind: SyntaxKind; pos: number }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly text: string }
interface PropertyAccessExpression extends Node { readonly kind: SyntaxKind.PropertyAccessExpression; flags: number; readonly name: Identifier }
interface ElementAccessExpression extends Node { readonly kind: SyntaxKind.ElementAccessExpression; flags: number; readonly index: number }
function mark(node: Node): void {
	(node as PropertyAccessExpression | ElementAccessExpression).flags = 5;
}
function move(node: Node): void {
	node.pos = 9;
}
function bump(node: Node): void {
	const access = node as PropertyAccessExpression | ElementAccessExpression;
	access.flags += 2;
	move(access);
}
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 1, text: 'a' };
const property: PropertyAccessExpression = { kind: SyntaxKind.PropertyAccessExpression, pos: 0, flags: 0, name: id };
const element: ElementAccessExpression = { kind: SyntaxKind.ElementAccessExpression, pos: 4, flags: 1, index: 2 };
const held: Node = element;
mark(property);
mark(held);
bump(element);
console.log(String(property.flags) + ' ' + String(element.flags) + ' ' + String(element.pos) + ' ' + String(property.pos));

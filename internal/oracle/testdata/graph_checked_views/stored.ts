interface Node { readonly kind: 'identifier' | 'number'; readonly pos: number; next: Node | undefined }
interface Identifier extends Node { readonly kind: 'identifier'; readonly name: string }
function identifier(node: Node): Identifier { return node as Identifier; }
interface Holder { node: Node | undefined; self: Holder | undefined }
function turn(index: number): string {
	const a: Identifier = { kind: 'identifier', pos: index, next: undefined, name: 'a' + String(index) };
	const seen: Node = a;
	const holder: Holder = { node: undefined, self: undefined };
	holder.self = holder;
	const view = identifier(seen);
	holder.node = view;
	return view.name + ' ' + String(holder.node === a);
}
let last = '';
for (let index = 0; index < 1000; index++) last = turn(index);
console.log(last);

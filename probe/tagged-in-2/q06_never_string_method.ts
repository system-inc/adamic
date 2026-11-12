interface Node { readonly kind: number }
interface TaggedNode extends Node { __tagged: undefined }
function tag(node: Node, sure: true): TaggedNode;
function tag(node: Node, sure: boolean): Node;
function tag(node: Node, sure: boolean): Node { return node; }
const r = tag({ kind: 1 }, true);
if (!('__tagged' in r)) {
	const lie: { kind: string } = r;
	console.log(lie.kind.toUpperCase());
}

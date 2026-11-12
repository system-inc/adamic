interface Node { readonly kind: number }
interface TaggedNode extends Node { __tagged: undefined }
function tag(node: Node, sure: true): TaggedNode;
function tag(node: Node, sure: boolean): Node;
function tag(node: Node, sure: boolean): Node { return node; }
const r = tag({ kind: 1 }, true);
const keep: Node = r;
if (!('__tagged' in r)) {
	const lie: { n: number; m: number } = r;
	lie.m = 7;
	console.log(String(lie.m));
	console.log(String(keep.kind));
}

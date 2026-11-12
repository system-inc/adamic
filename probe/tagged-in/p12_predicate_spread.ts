interface Node { readonly kind: number }
interface TaggedNode extends Node { __tagged: undefined }
function isTagged(node: Node): node is TaggedNode { return node.kind === 1; }
const n: Node = { kind: 1 };
if (isTagged(n)) {
	const copy = { ...n };
	console.log(String('__tagged' in copy));
	console.log(String(Object.keys(copy).length));
}

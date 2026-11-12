interface Node { readonly kind: number }
interface TaggedNode extends Node { __tagged: undefined }
function isTagged(node: Node): node is TaggedNode { return node.kind === 1; }
const n: Node = { kind: 1 };
if (isTagged(n)) {
	if (!('__tagged' in n)) {
		const lie: { kind: string } = n;
		console.log(String(lie.kind.length));
	}
}

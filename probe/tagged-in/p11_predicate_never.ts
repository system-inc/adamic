interface Node { readonly kind: number }
interface TaggedNode extends Node { __tagged: undefined }
function isTagged(node: Node): node is TaggedNode { return node.kind === 1; }
const n: Node = { kind: 1 };
if (isTagged(n)) {
	if ('__tagged' in n) {
		console.log('in');
	} else {
		const lie: string = n;
		console.log(typeof lie);
	}
}

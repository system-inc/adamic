interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
const child: N = { kind: 2, parent: root };
function kindOf(node: N): number {
	return node.parent?.kind;
}
console.log(`${kindOf(child)}`);
console.log(`${kindOf(root)}`);

interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
const child: N = { kind: 2, parent: root };
function kindOr(node: N): number {
	return node.parent?.kind || 0;
}
console.log(`${kindOr(child)}`);
console.log(`${kindOr(root)}`);

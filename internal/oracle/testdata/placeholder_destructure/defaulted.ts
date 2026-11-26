interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
const other: N = { kind: 2, parent: root };
function parentKind(node: N): string {
	const { parent = other } = node;
	return `${parent.kind}`;
}
console.log(parentKind(root));

interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
const child: N = { kind: 2, parent: root };
function show(k: number): string {
	return `${k}`;
}
function kindOf(node: N): string {
	return show(node.parent?.kind);
}
console.log(kindOf(child));
console.log(kindOf(root));

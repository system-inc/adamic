interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
const child: N = { kind: 2, parent: root };
function kinds(node: N): number[] {
	return [node.parent?.kind];
}
console.log(`${kinds(child)[0]}`);
console.log(`${kinds(root)[0]}`);

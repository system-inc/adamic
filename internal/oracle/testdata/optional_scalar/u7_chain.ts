interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
const child: N = { kind: 2, parent: root };
const grandchild: N = { kind: 3, parent: child };
function grand(node: N): string {
	return `${node.parent?.parent?.kind}`;
}
console.log(grand(grandchild));
console.log(grand(child));
console.log(grand(root));

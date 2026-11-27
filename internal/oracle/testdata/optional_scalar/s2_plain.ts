interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
const child: N = { kind: 2, parent: root };
function plain(node: N): string {
	return `${node.parent.kind}`;
}
console.log(plain(child));
console.log(plain(root));

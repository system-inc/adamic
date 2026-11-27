interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
const child: N = { kind: 2, parent: root };
function accessKind(node: N): string {
	let { parent } = node;
	switch (parent?.kind) {
		case 1:
			return "one";
		case 2:
			return "two";
		default:
			return "other";
	}
}
console.log(accessKind(child));
console.log(accessKind(root));

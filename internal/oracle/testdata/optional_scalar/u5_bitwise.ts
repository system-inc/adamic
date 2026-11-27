interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
const child: N = { kind: 2, parent: root };
function odd(node: N): string {
	if (node.parent?.kind & 1) {
		return "odd";
	}
	return "even or none";
}
console.log(odd(child));
console.log(odd(root));

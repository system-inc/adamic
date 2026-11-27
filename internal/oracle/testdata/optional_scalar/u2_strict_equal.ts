interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
const child: N = { kind: 2, parent: root };
function isOne(node: N): boolean {
	return node.parent?.kind === 1;
}
function isNotOne(node: N): boolean {
	return node.parent?.kind !== 1;
}
console.log(`${isOne(child)} ${isNotOne(child)}`);
console.log(`${isOne(root)} ${isNotOne(root)}`);

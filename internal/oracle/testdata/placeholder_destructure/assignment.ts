interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
function parentKind(node: N): string {
	let parent: N = node;
	({ parent } = node);
	return `${parent.kind}`;
}
console.log("before");
console.log(parentKind(root));
console.log("after");

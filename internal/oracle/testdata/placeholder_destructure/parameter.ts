interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
function parentKind({ parent }: N): string {
	return `${parent.kind}`;
}
console.log("before");
console.log(parentKind(root));
console.log("after");

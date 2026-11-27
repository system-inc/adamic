interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
const child: N = { kind: 2, parent: root };
interface Dump { readonly k: number | undefined; }
function dump(node: N): Dump {
	return { k: node.parent?.kind };
}
console.log(`${dump(child).k}`);
console.log(`${dump(root).k}`);

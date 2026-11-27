interface N { readonly kind: number; readonly parent: N; }
const root: N = { kind: 1, parent: undefined! };
const child: N = { kind: 2, parent: root };
interface Box { k: number; }
const box: Box = { k: 0 };
function store(node: N): void {
	box.k = node.parent?.kind;
}
store(child);
console.log(`${box.k}`);
store(root);
console.log(`${box.k}`);

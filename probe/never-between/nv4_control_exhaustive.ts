enum K { A = 1, B = 2 }
interface NodeA { readonly kind: K.A; readonly a: number }
interface NodeB { readonly kind: K.B; readonly b: number }
type N = NodeA | NodeB;
function stray(value: never): string { return "stray"; }
function describe(node: N): string {
	if (node.kind === K.A) return 'a';
	if (node.kind === K.B) return 'b';
	return stray(node);
}
console.log(describe({ kind: K.B, b: 2 }));
console.log(describe({ kind: K.A, a: 1 }));

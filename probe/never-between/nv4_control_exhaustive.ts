enum K { A = 1, B = 2 }
interface NodeA { readonly kind: K.A; readonly a: number }
interface NodeB { readonly kind: K.B; readonly b: number }
type N = NodeA | NodeB;
function stray(value: unknown): string { return "stray " + typeof value; }
function describe(node: N): string {
	if (node.kind === K.A) return 'a';
	if (node.kind === K.B) return 'b';
	{ const seen: unknown = node; return stray(seen); }
}
console.log(describe({ kind: K.B, b: 2 }));
console.log(describe({ kind: K.A, a: 1 }));

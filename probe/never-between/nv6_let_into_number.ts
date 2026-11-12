enum K { A = 1, B = 2 }
interface NodeA { readonly kind: K.A; readonly a: number }
interface NodeB { readonly kind: K.B; readonly b: number }
type N = NodeA | NodeB;
function stray(value: unknown): string { return "stray " + typeof value; }
function describe(start: N): string {
	let node: N = start;
	const flip = (): void => { node = { kind: K.A, a: 1 }; };
	if (node.kind === K.A) return 'a';
	flip();
	if (node.kind === K.B) return 'b';
	{ const count: number = node; return "count " + String(count + 1); }
}
console.log(describe({ kind: K.B, b: 2 }));

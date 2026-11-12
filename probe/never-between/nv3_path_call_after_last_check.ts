enum K { A = 1, B = 2 }
interface NodeA { readonly kind: K.A; readonly a: number }
interface NodeB { readonly kind: K.B; readonly b: number }
type N = NodeA | NodeB;
function stray(value: never): string { return "stray"; }
interface Holder { node: N }
function flip(holder: Holder): void { holder.node = { kind: K.A, a: 1 }; }
function describe(holder: Holder): string {
	if (holder.node.kind === K.A) return 'a';
	if (holder.node.kind === K.B) {
		flip(holder);
		if (holder.node.kind === K.B) return 'still b';
		return stray(holder.node);
	}
	return 'none';
}
console.log(describe({ node: { kind: K.B, b: 2 } }));

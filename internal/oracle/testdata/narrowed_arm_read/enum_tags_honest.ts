enum K {
	A,
	B,
}
type A = { readonly kind: K.A; readonly a: number };
type B = { readonly kind: K.B; readonly b: string };
type N = A | B;
interface M {
	n: N;
}
const a: A = { kind: K.A, a: 1 };
const other: A = { kind: K.A, a: 5 };
function reset(m: M): void {
	m.n = other;
}
function show(m: M): string {
	if (m.n.kind === K.A) {
		reset(m);
		return String(m.n.a + 1);
	}
	return "b";
}
console.log(show({ n: a }));

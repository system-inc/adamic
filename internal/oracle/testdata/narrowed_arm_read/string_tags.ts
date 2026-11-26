type A = { readonly kind: "a"; readonly a: number };
type B = { readonly kind: "b"; readonly b: string };
type N = A | B;
interface M {
	n: N;
}
const a: A = { kind: "a", a: 1 };
const b: B = { kind: "b", b: "x" };
function reset(m: M): void {
	m.n = b;
}
function show(m: M): string {
	if (m.n.kind === "a") {
		reset(m);
		return String(m.n.a + 1);
	}
	return "b";
}
console.log(show({ n: a }));

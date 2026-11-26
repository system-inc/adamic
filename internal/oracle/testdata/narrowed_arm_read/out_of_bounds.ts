type A = {
	readonly kind: "a";
	readonly first: number;
	readonly second: number;
	readonly third: number;
	readonly fourth: number;
	readonly fifth: number;
	readonly a: number;
};
type B = { readonly kind: "b"; readonly b: string };
type N = A | B;
interface M {
	n: N;
}
function reset(m: M): void {
	m.n = { kind: "b", b: "x".repeat(2) };
}
function show(m: M): string {
	if (m.n.kind === "a") {
		reset(m);
		return String(m.n.a + 1);
	}
	return "b";
}
console.log(show({ n: { kind: "a", first: 1, second: 2, third: 3, fourth: 4, fifth: 5, a: 6 } }));

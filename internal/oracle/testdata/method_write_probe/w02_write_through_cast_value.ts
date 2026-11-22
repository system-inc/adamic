type A = { readonly kind: 'a'; m(): number };
type B = { readonly kind: 'b'; m(): string };
function pick(n: number): A | B { return n > 0 ? { kind: 'a', m(): number { return 1; } } : { kind: 'b', m(): string { return 's'; } }; }
const u = pick(1);
const a = u as A;
a.m = (): number => 2;
console.log(String(a.m()));

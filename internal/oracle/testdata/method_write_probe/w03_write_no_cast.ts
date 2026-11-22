type A = { readonly kind: 'a'; m(): number };
const a: A = { kind: 'a', m(): number { return 1; } };
a.m = (): number => 2;
console.log(String(a.m()));

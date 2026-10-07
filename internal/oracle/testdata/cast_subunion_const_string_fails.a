const enum Kind { A = 'a', B = 'b', C = 'c' }
type A = { readonly kind: Kind.A; readonly n: number };
type B = { readonly kind: Kind.B; readonly n: number };
type C = { readonly kind: Kind.C; readonly n: number };
function cast(value: A | B | C): A | C { return value as A | C; }
console.log('casting');
console.log(`${cast({ kind: Kind.B, n: 2 }).n}`);

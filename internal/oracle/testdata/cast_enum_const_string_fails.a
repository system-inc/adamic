const enum Kind { A = 'a', B = 'b' }
type A = { readonly kind: Kind.A; readonly n: number };
type B = { readonly kind: Kind.B; readonly n: number };
function cast(value: A | B): A { return value as A; }
console.log('casting');
console.log(`${cast({ kind: Kind.B, n: 2 }).n}`);

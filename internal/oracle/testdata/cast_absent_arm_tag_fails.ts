// A present value whose tag the target doesn't hold, from a source with undefined beside it: the absent check
// passes and the tag check fails, with the same message. Node runs on without a check.
enum Kind { A = 1, B = 2, C = 3 }
type A = { readonly kind: Kind.A; readonly n: number };
type B = { readonly kind: Kind.B; readonly n: number };
type C = { readonly kind: Kind.C; readonly n: number };
function cast(value: A | B | C | undefined): A | C { return value as A | C; }
console.log('casting');
console.log(`${cast({ kind: Kind.B, n: 2 }).n}`);

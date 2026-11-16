// A lying undefined: the cast's target holds no undefined, so an undefined value fails it before any tag is
// read, with the cast's own message. Node runs on without a check and prints undefined.
enum Kind { A = 1, B = 2 }
type A = { readonly kind: Kind.A; readonly n: number };
type B = { readonly kind: Kind.B; readonly n: number };
function cast(value: A | B | undefined): A { return value as A; }
console.log('casting');
console.log(typeof cast(undefined));

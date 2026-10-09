enum K { A, B }
type A = { readonly kind: K.A; readonly n: number };
type B = { readonly kind: K.B; readonly n: number };
function narrow(value: A | B): A { return value as A; }

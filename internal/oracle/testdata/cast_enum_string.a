enum Kind { A = 'identifier', B = 'number', C = 'string' }
type A = { readonly kind: Kind.A; readonly text: string };
type B = { readonly kind: Kind.B; readonly text: string };
type C = { readonly kind: Kind.C; readonly text: string };
function one(node: A | B | C): A { return node as A; }
function several(node: A | B | C): A | C { return node as A | C; }
console.log(one({ kind: Kind.A, text: 'a'.repeat(3) }).text);
console.log(several({ kind: Kind.A, text: 'first'.repeat(2) }).text);
console.log(several({ kind: Kind.C, text: 'last'.repeat(2) }).text);

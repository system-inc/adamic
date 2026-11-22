// The earlier of two spreads is the one whose source may lack label. A second spread is refused before any
// field after it is judged (adamic/single-spread), so a field never follows more than one spread.
interface A { readonly id: number; label?: string }
interface B { readonly name: string }
const a: A = { id: 1 };
const b: B = { name: 'n' };
const c = { ...a, ...b, label: 'x' };
console.log(c.label); // Node: x

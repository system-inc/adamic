// label is optional in A, but the const a spreads is a fresh literal naming it, so a's object holds label and
// the label written after the spread overrides it where it stands. A literal spread in place is the same.
interface A { readonly id: number; label?: string }
const a: A = { id: 1, label: 'y' };
const c = { ...a, label: 'x' };
console.log(String(c.label)); // Node: x
console.log(Object.keys(c).join()); // Node: id,label
console.log(String(JSON.stringify(c))); // Node: {"id":1,"label":"x"}
const d = { ...{ id: 2, label: 'z' }, label: 'w' };
console.log(String(d.label)); // Node: w
console.log(Object.keys(d).join()); // Node: id,label

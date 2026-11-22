// The control: label is required in the source's type, so the source's object always holds it, and the label
// written after the spread overrides it where it stands.
interface A { readonly id: number; label: string }
const a: A = { id: 1, label: 'y' };
const c = { ...a, label: 'x' };
console.log(c.label); // Node: x
console.log(Object.keys(c).join()); // Node: id,label
console.log(JSON.stringify(c)); // Node: {"id":1,"label":"x"}

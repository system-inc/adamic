interface A { readonly id: number; label?: string }
const a: A = { id: 1 };
const c = { ...a, label: 'x' };
console.log(c.label);

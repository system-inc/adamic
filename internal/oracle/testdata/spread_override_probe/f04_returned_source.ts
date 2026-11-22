interface A { readonly id: number; label?: string }
function make(n: number): A { return n > 0 ? { id: 1 } : { id: 2, label: 'y' }; }
const a = make(1);
const c = { ...a, label: 'x' };
console.log(c.label);

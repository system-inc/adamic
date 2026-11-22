interface A { readonly id: number; label?: string }
interface B { readonly name: string }
const v: A & B = { id: 1, name: 'n' };
const c = { ...v, label: 'x' };
console.log(c.label);

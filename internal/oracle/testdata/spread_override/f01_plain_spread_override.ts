interface A { readonly id: number; valid?: boolean }
const a: A = { id: 1 };
const c = { ...a, valid: true };
console.log(String(c.valid));

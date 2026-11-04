interface Holder { readonly n: number; }
function reset(m: { n: number }): void { m.n = 2; }
const h: Holder = { n: 1 };
reset(h);
console.log(String(h.n));

interface Holder { readonly n: number; }
function reset<T extends { n: number }>(m: T): void { m.n = 2; }
const h: Holder = { n: 1 };
reset(h);
console.log(String(h.n));

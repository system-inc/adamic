interface R { readonly t: number }
interface W { t: number }
const r: R = { t: 2 };
const w: { p: W } = { get p(): R { return r; } };
w.p.t = 1;
console.log(String(r.t));

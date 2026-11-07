interface R { readonly t: number }
interface W { t: number }
const w: W = { t: 2 };
const o: { p: W } = { get p(): W { return w; } };
o.p.t = 1;
console.log(String(w.t));

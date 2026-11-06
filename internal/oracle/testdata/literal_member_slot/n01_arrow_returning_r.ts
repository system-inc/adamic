interface R { readonly t: number }
interface W { t: number }
const r: R = { t: 2 };
const f: () => W = (): R => r;
f().t = 1;
console.log(String(r.t));

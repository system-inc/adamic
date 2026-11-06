interface R { readonly t: number }
interface W { t: number }
const r: R = { t: 2 };
function make(): { p: W } | undefined {
    return { get p(): R { return r; } };
}
const x = make();
if (x) { x.p.t = 1; }
console.log(String(r.t));

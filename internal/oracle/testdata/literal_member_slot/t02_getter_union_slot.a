interface R { readonly t: number }
interface W { t: number }
const w: W = { t: 2 };
function make(): { p: W } | undefined {
    return { get p(): W { return w; } };
}
const x = make();
if (x) { x.p.t = 1; }
console.log(String(w.t));

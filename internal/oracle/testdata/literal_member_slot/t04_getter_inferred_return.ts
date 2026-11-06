interface R { readonly t: number }
interface W { t: number }
const w: W = { t: 2 };
function make(): { p: W } {
    return { get p() { return w; } };
}
make().p.t = 1;
console.log(String(w.t));

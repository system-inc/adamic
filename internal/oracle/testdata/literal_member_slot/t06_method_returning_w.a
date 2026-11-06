interface R { readonly t: number }
interface W { t: number }
const w: W = { t: 2 };
function make(): { get(): W } {
    return { get(): W { return w; } };
}
make().get().t = 1;
console.log(String(w.t));

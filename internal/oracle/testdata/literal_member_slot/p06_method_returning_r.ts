interface R { readonly t: number }
interface W { t: number }
const r: R = { t: 2 };
function make(): { get(): W } {
    return { get(): R { return r; } };
}
make().get().t = 1;
console.log(String(r.t));

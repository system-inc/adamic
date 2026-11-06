interface R { readonly t: number }
interface W { t: number }
const r: R = { t: 2 };
function make(): { p: W } {
    return { get p(): R { return r; } };
}
make().p.t = 1;
console.log(String(r.t));

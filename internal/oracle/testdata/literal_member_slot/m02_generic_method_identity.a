interface R { readonly t: number }
interface W { t: number }
const r: R = { t: 2 };
function make(): { get(x: R): W } {
    return { get<T>(x: T): T { return x; } };
}
make().get(r).t = 1;
console.log(String(r.t));

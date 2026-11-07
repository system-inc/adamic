interface R { readonly t: number }
interface W { t: number }
const r: R = { t: 2 };
function make(k: 'x' | 'y'): { k: 'x'; get(): W } | { k: 'y'; get(): W } {
    return { k, get(): R { return r; } };
}
make('x').get().t = 1;
console.log(String(r.t));

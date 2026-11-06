interface R { readonly t: number }
interface W { t: number }
const w: W = { t: 2 };
function use(o: { p: W }): void {
    o.p.t = 1;
}
use({ get p(): W { return w; } });
console.log(String(w.t));

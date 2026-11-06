interface R { readonly t: number }
interface W { t: number }
const r: R = { t: 2 };
function use(o: { p: W }): void {
    o.p.t = 1;
}
use({ get p(): R { return r; } });
console.log(String(r.t));

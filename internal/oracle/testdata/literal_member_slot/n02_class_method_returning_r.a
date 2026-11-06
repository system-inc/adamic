interface R { readonly t: number }
interface W { t: number }
const r: R = { t: 2 };
class Box {
    get(): R { return r; }
}
const b: { get(): W } = new Box();
b.get().t = 1;
console.log(String(r.t));

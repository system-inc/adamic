// #4gpbn7w: the same dispatcher read compared with undefined and called when present, over a getter
// owner, a literal without the field and a literal with it, in a loop so an out-of-bounds read can't
// pass by luck on one allocation.
interface H {
    readonly m?: () => number;
}
class G implements H {
    get m(): () => number {
        return (): number => 7;
    }
}
function value(h: H): number {
    const m = h.m;
    return m === undefined ? -1 : m();
}
let total = 0;
for (let i = 0; i < 100; i++) {
    total += value(new G()) + value({}) + value({ m: (): number => i });
}
console.log(String(total));

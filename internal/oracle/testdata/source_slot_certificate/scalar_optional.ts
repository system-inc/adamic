// #f95gkw2: the same write into a field both types declare optional, through a view that adds a required id,
// so note's presence is proven compatible and the write is the only stop.
interface Base { readonly kind: number; note?: string | undefined; }
interface Nd extends Base { readonly kind: number; readonly id: number; note?: string | undefined; }
const raw = { kind: 1, id: 7, note: 'a' as string | undefined };
const base: Base = raw;
const n = base as Nd;
n.note = undefined;
console.log(n.note ?? 'none');

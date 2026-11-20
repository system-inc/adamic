// #f95gkw2: the same write into a field the view declares optional.
interface Base { readonly kind: number; }
interface Nd extends Base { readonly kind: number; note?: string | undefined; }
const raw = { kind: 1, note: 'a' };
const base: Base = raw;
const n = base as Nd;
n.note = undefined;
console.log(n.note ?? 'none');

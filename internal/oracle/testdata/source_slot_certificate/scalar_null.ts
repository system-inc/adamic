// #f95gkw2: null into a string | null field a literal filled with a string.
interface Base { readonly kind: number; }
interface Nd extends Base { readonly kind: number; note: string | null; }
const raw = { kind: 1, note: 'a' as string | null };
const base: Base = raw;
const n = base as Nd;
n.note = null;
console.log(n.note ?? 'none');

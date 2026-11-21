// #f95gkw2: a number | undefined field a literal filled with a plain number, which can't hold undefined.
interface Base { readonly kind: number; }
interface Nd extends Base { readonly kind: number; count: number | undefined; }
const raw = { kind: 1, count: 2 as number | undefined };
const base: Base = raw;
const n = base as Nd;
n.count = undefined;
console.log(String(n.count ?? -1));

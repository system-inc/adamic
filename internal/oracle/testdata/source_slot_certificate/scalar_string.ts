// #f95gkw2: the twin of scalar_undefined.ts writing a string, which the slot already holds, so it writes as before.
interface Base { readonly kind: number; }
interface Nd extends Base { readonly kind: number; note: string | undefined; }
const raw = { kind: 1, note: 'a' as string | undefined };
const base: Base = raw;
const n = base as Nd;
n.note = 'b';
console.log(n.note ?? 'none');

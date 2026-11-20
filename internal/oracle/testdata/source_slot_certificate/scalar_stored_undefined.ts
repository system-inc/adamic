// #f95gkw2: a literal stored undefined, so this write of undefined passes the checked write today. What the slot
// holds is set by every write before this one (the 'c' below changes it), so no compile-time proof of it is asked,
// and the write fails closed with the rest.
interface Base { readonly kind: number; }
interface Nd extends Base { readonly kind: number; note: string | undefined; }
const raw = { kind: 1, note: undefined as string | undefined };
const base: Base = raw;
const n = base as Nd;
n.note = 'c';
n.note = undefined;
console.log(n.note ?? 'none');

// #f95gkw2, Rule C's scalar arm: a checked view writes undefined into a string | undefined field a literal filled
// with a string. The checked write can't change the slot from a string to undefined, so it stopped at exit 70
// where Node runs (J-hfx8z13 at main 49ecfe75), and the write is NotYet.
interface Base { readonly kind: number; }
interface Nd extends Base { readonly kind: number; parent?: Nd | undefined; note: string | undefined; }
const raw = { kind: 1, note: 'a' as string | undefined };
const base: Base = raw;
const n = base as Nd;
n.note = undefined;
console.log(String(n.kind));

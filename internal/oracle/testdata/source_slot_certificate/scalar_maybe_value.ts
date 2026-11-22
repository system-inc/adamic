// #f95gkw2: a value that may be undefined but isn't the literal passes the checked write as the slot's own
// representation, and raw, which declares note a string, then read a null string natively: a segfault where
// Node throws (J-hfx8z13, exit 139). Rule C's scalar arm stops the write.
interface Base { readonly kind: number; }
interface Nd extends Base { readonly kind: number; note: string | undefined; }
function pick(flag: boolean): string | undefined { return flag ? 'x' : undefined; }
const raw = { kind: 1, note: 'a' };
const base: Base = raw;
const n = base as Nd;
n.note = pick(false);
console.log(typeof raw.note);

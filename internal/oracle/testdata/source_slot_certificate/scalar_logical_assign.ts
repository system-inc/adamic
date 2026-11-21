// #f95gkw2: ||= can store its right side, which may be undefined.
interface Base { readonly kind: number; }
interface Nd extends Base { readonly kind: number; note: string | undefined; }
function pick(flag: boolean): string | undefined { return flag ? 'x' : undefined; }
const raw = { kind: 1, note: '' as string | undefined };
const base: Base = raw;
const n = base as Nd;
n.note ||= pick(false);
console.log(String(n.kind));

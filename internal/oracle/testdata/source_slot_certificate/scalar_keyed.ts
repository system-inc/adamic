// #f95gkw2: the same write by a literal key, judged by the keyed write's guard as the named write is.
interface Base { readonly kind: number; }
interface Nd extends Base { readonly kind: number; note: string | undefined; }
const raw = { kind: 1, note: 'a' as string | undefined };
const base: Base = raw;
const n = base as Nd;
const key = 'note' as const;
n[key] = undefined;
console.log(String(n.kind));

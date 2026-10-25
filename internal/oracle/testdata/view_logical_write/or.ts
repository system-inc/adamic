// ||= through the same view, over a count of 0: the write is taken, and boxed into the union slot.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
function initial(): string | number { return 0; }
const raw = { kind: 'Holder' as const, count: initial() };
const base: Root = raw;
const viewed = base as View;
const result = (viewed.count ||= 7);
const read: string | number = raw.count;
console.log(typeof read === 'number' ? `number ${read + 1} ${result}` : `string ${read}`);

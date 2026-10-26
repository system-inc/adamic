// ||= through the object's own type, with a view of it elsewhere: the write is to string | number,
// boxed as any union write is, so it still compiles and raw.count reads the number back.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
function initial(): string | number { return ''; }
const raw = { kind: 'Holder' as const, count: initial() };
const base: Root = raw;
const viewed = base as View;
raw.count ||= 7;
const read: string | number = raw.count;
console.log(typeof read === 'number' ? `number ${read + 1}` : `string ${read}`);
console.log(viewed.kind);

// A number written through a checked view into a field stored as string | number: the write is
// checked and boxed, as a read through the view is, so raw.count reads the number back.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
function initial(): string | number { return 'abc'; }
const raw = { kind: 'Holder' as const, count: initial() };
const base: Root = raw;
const viewed = base as View;
viewed.count = 7;
const read: string | number = raw.count;
console.log(typeof read === 'number' ? `number ${read + 1}` : `string ${read}`);

// A view claims count is a number over an object whose count is a string. Node writes 7 and the
// string-typed read then finds a number; Adamic's view write check stops at the write instead.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
const raw = { kind: 'Holder' as const, count: 'abc' };
const base: Root = raw;
const viewed = base as View;
viewed.count = 7;
const read: string = raw.count;
console.log(`${typeof read} ${read.length}`);

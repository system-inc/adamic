// A number written through a checked view into a field its producer's literal stores as a narrowed
// string. The slot's runtime type says string, which a string field says too, so the view write check
// refuses it rather than box a number where a string may be read. Before the write carried its
// receiver's key this was a plain number store, and raw.count read the double as a reference.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
const start: string | number = 'abc';
const raw = { kind: 'Holder' as const, count: start as string | number };
const base: Root = raw;
const viewed = base as View;
viewed.count = 7;
const read: string | number = raw.count;
console.log(typeof read === 'number' ? `number ${read + 1}` : `string ${read}`);

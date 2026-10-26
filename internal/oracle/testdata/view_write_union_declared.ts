// A number written through a checked view into a field whose producer declares it string | number:
// the literal stores the string as a union, so the write is checked and boxed and raw.count reads
// the number back.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
const raw: { readonly kind: 'Holder'; count: string | number } = { kind: 'Holder', count: 'abc' };
const base: Root = raw;
const viewed = base as View;
viewed.count = 7;
const read: string | number = raw.count;
console.log(typeof read === 'number' ? `number ${read + 1}` : `string ${read}`);

// ??= through a view whose count is a number never writes, but still reads count through the view.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
function initial(): string | number { return 5; }
const raw = { kind: 'Holder' as const, count: initial() };
const base: Root = raw;
const viewed = base as View;
const result = (viewed.count ??= 7);
const read: string | number = raw.count;
console.log(typeof read === 'number' ? `number ${read + 1} ${result}` : `string ${read}`);

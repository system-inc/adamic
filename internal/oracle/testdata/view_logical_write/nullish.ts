// ??= through a view whose count is number | null, over a slot holding string | number | null: null is
// its own box in a union, and ??= takes the write over it as ?? would.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number | null; }
function initial(): string | number | null { return null; }
const raw = { kind: 'Holder' as const, count: initial() };
const base: Root = raw;
const viewed = base as View;
const result = (viewed.count ??= 7);
const read: string | number | null = raw.count;
console.log(typeof read === 'number' ? `number ${read + 1} ${result}` : read === null ? 'null' : `string ${read}`);

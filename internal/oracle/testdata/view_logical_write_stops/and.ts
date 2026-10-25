// &&= through a view whose count is a number, over an object holding the string 'abc' in count: the
// checked read stops, as read.ts's plain read does, where Node prints "number 8".
// A ruled negative witness (#995z2fa): and_repaired.ts returns 5 in place of 'abc' and agrees with Node.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
function initial(): string | number { return 'abc'; }
const raw = { kind: 'Holder' as const, count: initial() };
const base: Root = raw;
const viewed = base as View;
viewed.count &&= 7;
const read: string | number = raw.count;
console.log(typeof read === 'number' ? `number ${read + 1}` : `string ${read}`);

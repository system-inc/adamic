// &&= of a boolean through a view, over a string | boolean slot: false is boxed into the union.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { flag: boolean; }
function initial(): string | boolean { return true; }
const raw = { kind: 'Holder' as const, flag: initial() };
const base: Root = raw;
const viewed = base as View;
viewed.flag &&= false;
const read: string | boolean = raw.flag;
console.log(typeof read === 'boolean' ? `boolean ${read}` : `string ${read}`);

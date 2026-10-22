// A boolean written through a checked view into a field stored as boolean | string: the write is
// checked and given its constant box, so raw.on reads the boolean back.
interface Root { readonly kind: 'Flag'; }
interface View extends Root { on: boolean; }
function initial(): boolean | string { return 'off'; }
const raw = { kind: 'Flag' as const, on: initial() };
const base: Root = raw;
const viewed = base as View;
viewed.on = true;
const read: boolean | string = raw.on;
console.log(typeof read === 'boolean' ? `boolean ${read}` : `string ${read}`);

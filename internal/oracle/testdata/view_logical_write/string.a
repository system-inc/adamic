// ||= and &&= of strings through a view, over a string | number slot: the reference written is kept by
// the slot, and the old one let go, as a plain = through the view does.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { name: string; }
function initial(): string | number { return ''; }
const raw = { kind: 'Holder' as const, name: initial() };
const base: Root = raw;
const viewed = base as View;
const result = (viewed.name ||= 'filled');
viewed.name &&= `${viewed.name} twice`;
const read: string | number = raw.name;
console.log(typeof read === 'number' ? `number ${read + 1}` : `string ${read} ${result}`);

// A good view whose string field is stored by its producer as string | number.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly name: string; }
const name: string | number = 'abc';
const raw = { kind: 'Holder' as const, name: name as string | number };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.name.length}`);
raw.name = 5;
console.log(`${viewed.name.length}`);

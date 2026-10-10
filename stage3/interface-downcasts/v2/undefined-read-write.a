interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: undefined | string; }
const raw = { kind: 'Holder' as const, value: undefined };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.value === undefined}`);

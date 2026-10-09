interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: boolean | undefined; }
const raw = { kind: 'Holder' as const, value: 'wrong' };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.value}`);

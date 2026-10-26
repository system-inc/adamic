interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: number | undefined; }
interface Holder { readonly kind: 'Holder'; value: number | undefined; }
const raw: Holder = { kind: 'Holder', value: undefined };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.value}`);
raw.value = 0;
console.log(`${viewed.value}`);

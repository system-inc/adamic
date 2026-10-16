interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: boolean | undefined; }
interface Holder { readonly kind: 'Holder'; value: boolean | undefined; }
const raw: Holder = { kind: 'Holder', value: undefined };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.value}`);
raw.value = false;
console.log(`${viewed.value}`);

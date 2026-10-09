interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: string | undefined; }
interface Holder { readonly kind: 'Holder'; value: string | undefined; }
const raw: Holder = { kind: 'Holder', value: undefined };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.value}`);
raw.value = ''.concat('x');
console.log(`${viewed.value}`);

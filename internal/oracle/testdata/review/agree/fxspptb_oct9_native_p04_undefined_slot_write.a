interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: undefined | string; }
interface Holder { readonly kind: 'Holder'; value: string | undefined; }
const raw: Holder = { kind: 'Holder', value: undefined };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.value === undefined}`);
raw.value = 'written'.repeat(2);
console.log(`${viewed.value}`);
raw.value = undefined;
console.log(`${viewed.value === undefined}`);

interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: number | string; }
interface Other { value: string | undefined; }
const raw = { kind: 'Holder' as const, value: 3 };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.value}`);
const other: Other = { value: 'x'.repeat(3) };
console.log(`${other.value}`);
other.value = undefined;
console.log(`${other.value === undefined}`);

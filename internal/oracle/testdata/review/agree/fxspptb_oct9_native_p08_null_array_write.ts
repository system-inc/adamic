interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: number | string; }
interface Other { value: number[] | null; items: Map<string, number> | null; }
const raw = { kind: 'Holder' as const, value: 3 };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.value}`);
const other: Other = { value: null, items: null };
other.value = [1, 2];
console.log(`${other.value === null ? -1 : other.value.length}`);
other.value = null;
console.log(`${other.value === null}`);

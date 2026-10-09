interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: number | string; }
interface Other { readonly value: string | undefined; }
const raw = { kind: 'Holder' as const, value: 3 };
const base: Root = raw;
console.log(`${(base as View).value}`);
const other: Other = { value: undefined };
console.log(`${other.value}`);

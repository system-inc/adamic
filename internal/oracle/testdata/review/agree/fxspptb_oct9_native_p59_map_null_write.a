interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly table: number | string; }
interface Other { table: Map<string, number> | null; }
const raw = { kind: 'Holder' as const, table: 3 };
const base: Root = raw;
console.log(`${(base as View).table}`);
const other: Other = { table: null };
other.table = new Map<string, number>();
console.log(`${other.table === null ? -1 : other.table.size}`);

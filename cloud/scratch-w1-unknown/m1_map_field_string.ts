interface Holder { readonly table: Map<string, number> }
const h: Holder = { table: new Map<string, number>([['a', 1]]) };
const u: unknown = h;
if (typeof u === 'object' && u !== null && 'table' in u) {
	console.log(String(u.table));
}

interface Holder { readonly table: Map<string, number> }
const h: Holder = { table: new Map<string, number>([['a', 1]]) };
const u: unknown = h;
if (typeof u === 'object' && u !== null && 'table' in u) {
	const t = u.table;
	if (typeof t === 'object' && t !== null) {
		console.log(String('size' in t));
	}
}

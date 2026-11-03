// roads/generic.ts's road, with the record holding 37 under a field typed as the whole numeric enum Kind.
// Numeric enums are open (docs/rulings.md, resolved October 19), so 37 is a Kind, and the read agrees
// with Node.
enum Kind { A = 1, B = 99, C = 5 }
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["kind"] = 37;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly kind?: Kind }
function read<T extends Options>(t: T): string {
	return String(t.kind);
}
console.log(read(table));

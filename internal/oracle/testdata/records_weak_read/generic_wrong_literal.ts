// roads/generic.ts's road, with the record holding Kind.C, 5, under a field typed Kind.A | Kind.B, a
// strict subset of Kind's members. typeof says number; the literal membership stops it, naming the 5.
enum Kind { A = 1, B = 99, C = 5 }
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["kind"] = Kind.C;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly kind?: Kind.A | Kind.B }
function read<T extends Options>(t: T): string {
	return String(t.kind);
}
console.log(read(table));

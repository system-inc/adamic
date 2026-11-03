// roads/generic.ts's road, with the record holding a string under a field typed 1 | 2. Node prints the
// string; natively and on the JavaScript backend the read stops, naming it, where read as 1 | 2 the
// string would be taken for a number.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["mode"] = "fast";
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly mode?: 1 | 2 }
function read<T extends Options>(t: T): string {
	return String(t.mode);
}
console.log(read(table));

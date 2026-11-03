// A record holding a string under a field typed 1 | 2, seen as the weak type: tsc admits it, and Node
// prints the string. Natively and on the JavaScript backend the read stops, naming the field, where
// read as 1 | 2 the string would be taken for a number.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["mode"] = "fast";
interface Options { readonly total?: number; readonly mode?: 1 | 2 }
function mode(options: Options): string {
	return String(options.mode);
}
for (const key of Object.keys(table)) console.log(key);
console.log(mode({ mode: 2 }));
console.log(mode(table));

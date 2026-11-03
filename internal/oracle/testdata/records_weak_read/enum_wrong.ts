// enum_union.ts's field, Kind.A | Kind.B, with a laundered record holding Kind.C, a number the field's
// type doesn't have. Node prints 5; natively and on the JavaScript backend the read stops, naming it.
enum Kind { A = 1, B = 99, C = 5 }
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = Kind.C;
interface Options { readonly total?: number; readonly overrideImportMode?: Kind.A | Kind.B }
function mode(options: Options): string {
	return String(options.overrideImportMode);
}
for (const key of Object.keys(table)) console.log(key);
console.log(mode({ total: 1, overrideImportMode: Kind.B }));
console.log(mode(table));

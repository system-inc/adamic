// moduleSpecifiers.ts's field type: a union of two of an enum's members, ModuleKind.ESNext |
// ModuleKind.CommonJS there. The record's value is held to the whole union.
enum Kind { A = 1, B = 99, C = 5 }
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = Kind.B;
interface Options { readonly total?: number; readonly overrideImportMode?: Kind.A | Kind.B }
function mode(options: Options): string {
	return String(options.overrideImportMode);
}
for (const key of Object.keys(table)) console.log(key);
console.log(mode({ total: 1, overrideImportMode: Kind.A }));
console.log(mode(table));

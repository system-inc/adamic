// A road for a record into a weak read: a union Options | number holding the record.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = 1;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly overrideImportMode?: 1 | 99 }
function mode(options: Options): string {
	return String(options.overrideImportMode);
}
function pick(value: Options | number): string {
	return typeof value === "number" ? String(value) : mode(value);
}
console.log(pick(table));

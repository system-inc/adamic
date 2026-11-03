// A road for a record into a weak read: the record returned from a function declared to return the
// weak type.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = 1;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly overrideImportMode?: 1 | 99 }
function mode(options: Options): string {
	return String(options.overrideImportMode);
}
function asOptions(): Options {
	return table;
}
console.log(mode(asOptions()));

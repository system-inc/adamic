// A road for a record into a weak read: the record stored in a field typed Options, then read back.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = 1;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly overrideImportMode?: 1 | 99 }
function mode(options: Options): string {
	return String(options.overrideImportMode);
}
interface Holder { readonly options: Options }
const holder: Holder = { options: table };
console.log(mode(holder.options));

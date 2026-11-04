// A road for a record into a weak read: a readonly Table[] seen as a readonly Options[].
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = 1;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly overrideImportMode?: 1 | 99 }
function mode(options: Options): string {
	return String(options.overrideImportMode);
}
const tables: readonly Table[] = [table];
const list: readonly Options[] = tables;
for (const options of list) console.log(mode(options));

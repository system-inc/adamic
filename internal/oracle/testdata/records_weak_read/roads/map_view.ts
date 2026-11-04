// A road for a record into a weak read: a ReadonlyMap<string, Table> seen as a ReadonlyMap<string,
// Options>, then a value read out of it.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = 1;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly overrideImportMode?: 1 | 99 }
function mode(options: Options): string {
	return String(options.overrideImportMode);
}
const tables: ReadonlyMap<string, Table> = new Map<string, Table>([["a", table]]);
const views: ReadonlyMap<string, Options> = tables;
const found = views.get("a");
if (found !== undefined) console.log(mode(found));

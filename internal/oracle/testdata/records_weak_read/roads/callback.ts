// A road for a record into a weak read: a callback typed (row: Table) => string, handed a function of
// (o: Options) => string, the contravariant direction.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = 1;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly overrideImportMode?: 1 | 99 }
function mode(options: Options): string {
	return String(options.overrideImportMode);
}
function each(rows: readonly Table[], show: (row: Table) => string): void {
	for (const row of rows) console.log(show(row));
}
each([table], mode);

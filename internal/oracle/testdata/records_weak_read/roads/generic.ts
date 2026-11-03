// A road for a record into a weak read: a type parameter constrained by the weak type, instantiated
// with the record's type.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = 1;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly overrideImportMode?: 1 | 99 }
function read<T extends Options>(t: T): string {
	return String(t.overrideImportMode);
}
console.log(read(table));

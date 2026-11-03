// The road that reaches the record path (roads/generic.ts): a type parameter constrained by the weak
// type, instantiated with the record's type. The read takes the record branch and reads the record's
// key; a key the record never wrote reads undefined, as Node reads it.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = 1;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly overrideImportMode?: 1 | 99; readonly preference?: number }
function read<T extends Options>(t: T): string {
	return String(t.overrideImportMode);
}
function preference<T extends Options>(t: T): string {
	return String(t.preference);
}
console.log(read(table));
console.log(preference(table));

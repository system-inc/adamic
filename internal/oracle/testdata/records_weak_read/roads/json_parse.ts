// A road for a record into a weak read: JSON's own parse, whose any is cast to the weak type.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = 1;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly overrideImportMode?: 1 | 99 }
function mode(options: Options): string {
	return String(options.overrideImportMode);
}
const text: string = '{"total":1,"overrideImportMode":99}';
console.log(mode(JSON.parse(text) as Options));

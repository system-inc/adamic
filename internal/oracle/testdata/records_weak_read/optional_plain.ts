// moduleSpecifiers.ts's shape: options.overrideImportMode, an optional field of a weak type (every
// field optional), read plainly. Table is record storage and assignable to Options, so a record can
// be the receiver: the read tests its storage as it runs. A field the record never wrote reads
// undefined, as an absent field does.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = 1;
interface Options { readonly total?: number; readonly overrideImportMode?: 1 | 99; readonly preference?: number }
function mode(options: Options): string {
	return String(options.overrideImportMode);
}
function preference(options: Options): string {
	return String(options.preference);
}
for (const key of Object.keys(table)) console.log(key);
console.log(mode({ total: 1, overrideImportMode: 99 }));
console.log(mode(table));
console.log(preference({ preference: 4 }));
console.log(preference(table));

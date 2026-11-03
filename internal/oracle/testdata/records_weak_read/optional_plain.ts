// moduleSpecifiers.ts's shape: options.overrideImportMode, an optional field of a weak type (every
// field optional), read plainly. Table is record storage and its type is assignable to Options, so as
// far as the types say a record may be the receiver, and the read tests its storage as it runs. No
// program hands it one (Adamic refuses a record seen as the weak type at the relation), so only fixed
// objects are read here; TestWeakReadRecordPath proves the record path below the language.
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
console.log(mode({ total: 1 }));
console.log(preference({ preference: 4 }));

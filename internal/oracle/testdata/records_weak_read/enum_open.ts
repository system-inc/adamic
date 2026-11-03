// A field typed as a whole numeric enum, Kind. Numeric enums are open (docs/rulings.md, resolved
// October 19), so any number is a Kind and the read holds a record's value to number alone; a strict
// subset of Kind's members, as enum_union.ts's, is held member by member. Fixed objects only, as in
// optional_plain.ts; TestWeakReadRecordPath reads a record's 37 through such a field.
enum Kind { A = 1, B = 99, C = 5 }
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = 37;
interface Options { readonly total?: number; readonly overrideImportMode?: Kind }
function mode(options: Options): string {
	return String(options.overrideImportMode);
}
for (const key of Object.keys(table)) console.log(key);
console.log(mode({ total: 1, overrideImportMode: Kind.C }));

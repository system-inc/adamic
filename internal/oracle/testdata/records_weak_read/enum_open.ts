// A field typed as a whole numeric enum, Kind, whose laundered record holds 37, a number none of Kind's
// members is. Numeric enums are open (docs/rulings.md, resolved October 19), so 37 is a Kind: the read
// holds it to number only and agrees with Node. enum_wrong.ts's Kind.A | Kind.B, a strict subset of
// Kind's members, still stops on Kind.C.
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
console.log(mode(table));

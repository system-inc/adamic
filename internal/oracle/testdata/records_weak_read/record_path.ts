// The record path's proof below the language (TestWeakReadRecordPath). Adamic refuses a record seen as a
// weak type at the relation, since a weak type's optional field the record's type lacks could hide one,
// and #xywvn06 closes the one road past it, a generic's type argument. So the test lowers this
// fixed-receiver program and rewrites one call's argument in the IR, to a record table (an entries
// record literal) or to the class instance settings. As written, every read is a fixed object's, and the
// program agrees with Node.
enum Kind { A = 1, B = 99, C = 5 }
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["mode"] = 1;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly mode?: 1 | 2; readonly kind?: Kind.A | Kind.B; readonly open?: Kind }
class Settings {
	total = 3;
	mode: 1 | 2 = 2;
}
const settings = new Settings();
console.log(String(settings.total));
function readMode(options: Options): string {
	return String(options.mode);
}
function readKind(options: Options): string {
	return String(options.kind);
}
function readOpen(options: Options): string {
	return String(options.open);
}
console.log(readMode({ mode: 2 }));
console.log(readKind({ kind: Kind.A }));
console.log(readOpen({ open: Kind.C }));

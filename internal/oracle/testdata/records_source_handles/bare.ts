// #ny0c3gb's bare form: a record literal no place holds as anything more specific than itself, so it's
// held as its own widened type, `{}`, which every object type relates to. entriesRecordLiteral admits only
// a const's initializer, so the `{}` is bound with no annotation and passed straight to a record-typed
// parameter, and an annotated alias writes it. Both reads stay checked: Options's, which a record held as
// Record<string, number> couldn't reach (handle.ts), and Hints's, a weak type any record reaches.
interface Options { readonly label: string; readonly mode?: 1 | 99 }
interface Hints { readonly depth?: 1 | 2 }
const table = {};
const writer: Record<string, number> = table;
writer["alpha"] = 1;
writer["beta"] = 2;
for (const key of Object.keys(table)) console.log(key);
function count(values: Record<string, number>): number {
	return Object.keys(values).length;
}
console.log(String(count(table)));
function mode(options: Options): string {
	return String(options.mode);
}
function depth(hints: Hints): string {
	return String(hints.depth);
}
console.log(mode({ label: "a", mode: 99 }));
console.log(mode({ label: "b" }));
console.log(depth({ depth: 2 }));
console.log(depth({}));

// #ny0c3gb's satisfies form: `satisfies T` checks the literal against T and leaves it the type it had
// (rulings.md:14), so T holds nothing, and the record is held as its own widened type, `{}`. Both reads
// stay checked. Counting T as a holding place (record-source-handles.mutants.md) leaves Options's read
// unchecked, and this goes red.
interface Options { readonly label: string; readonly mode?: 1 | 99 }
interface Hints { readonly depth?: 1 | 2 }
const table = {} satisfies Record<string, number>;
const writer: Record<string, number> = table;
writer["alpha"] = 1;
writer["beta"] = 2;
for (const key of Object.keys(table)) console.log(key);
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

// #ny0c3gb's control: a record literal with a real handle, a const declared Record<string, number>, is
// held as that type alone. Options declares a required label, so no Record<string, number> relates to it
// either way, and its read is plain. Hints is a weak type a Record<string, number> is assignable to, so
// its read stays checked, which shows the program has its record.
interface Options { readonly label: string; readonly mode?: 1 | 99 }
interface Hints { readonly depth?: 1 | 2 }
const table: Record<string, number> = {};
table["alpha"] = 1;
table["beta"] = 2;
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

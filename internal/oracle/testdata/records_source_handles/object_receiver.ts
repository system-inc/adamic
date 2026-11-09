// #ny0c3gb: a receiver with an `object` member. The `{}` question isn't asked of the non-primitive object member,
// and the sources loop answers for it: the record's handle, Record<string, number>, is assignable to object, so the
// read of mode through `object & Options` stays checked. Options alone would read its slot (handle.ts).
interface Options { readonly label: string; readonly mode?: 1 | 99 }
const table: Record<string, number> = {};
table["alpha"] = 1;
for (const key of Object.keys(table)) console.log(key);
function mode(options: object & Options): string {
	return String(options.mode);
}
console.log(mode({ label: "a", mode: 99 }));
console.log(mode({ label: "b" }));

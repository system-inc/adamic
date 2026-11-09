// #ny0c3gb: a receiver with a structural object member. Receiver is all-optional, so the implicit `{}`
// source reaches it and keeps the read through Receiver & Options checked. Options alone would read its
// slot (handle.ts). A non-primitive `object` member instead routes the read through boxed dynamic storage.
interface Receiver { readonly hint?: number }
interface Options { readonly label: string; readonly mode?: 1 | 99 }
const table: Record<string, number> = {};
table["alpha"] = 1;
for (const key of Object.keys(table)) console.log(key);
function mode(options: Receiver & Options): string {
	return String(options.mode);
}
console.log(mode({ label: "a", mode: 99 }));
console.log(mode({ label: "b" }));

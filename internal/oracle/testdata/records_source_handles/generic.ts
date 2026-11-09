// #ny0c3gb's generic form: a record handed to keep<T> is held as the T inferred for it, its own `{}`. A
// literal written straight into keep's argument is no record by entriesRecordLiteral's test, which admits
// only a const's initializer, so the `{}` is bound with no annotation first and handed to keep by name.
// Both reads stay checked.
interface Options { readonly label: string; readonly mode?: 1 | 99 }
interface Hints { readonly depth?: 1 | 2 }
function keep<T>(value: T): T {
	return value;
}
const table = {};
const writer: Record<string, number> = table;
writer["alpha"] = 1;
writer["beta"] = 2;
for (const key of Object.keys(table)) console.log(key);
function count(values: Record<string, number>): number {
	return Object.keys(values).length;
}
console.log(String(count(keep(table))));
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

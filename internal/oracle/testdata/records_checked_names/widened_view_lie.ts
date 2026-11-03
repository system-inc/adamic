// A record read through a widened view of its own type stays checked: Table is assignable to Seen, so
// a record can be a Seen and its receiver is never plain. The write puts a string under total, typed
// number, through a computed key, and the read stops on it; Node prints "seven1".
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["to" + "tal"] = "seven";
for (const key of Object.keys(table)) console.log(key);
interface Seen { readonly total: number }
function show(seen: Seen): string {
	return String(seen.total + 1);
}
console.log(show(table));

// #9jnjwrw's crash: a record writes mode, so readiness checked every read named mode, and a class
// instance read through a generic instantiated with its class, an ordinary optional number read, was
// emitted as a checked view field, which can't be one ("incomplete checked field contract"). No record
// and no merged class can be a Settings, so its receiver is plain and the read is a slot read.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["mode"] = 1;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly mode?: 1 | 2 }
class Settings {
	total = 3;
	mode: 1 | 2 = 2;
}
function read<T extends Options>(t: T): string {
	return String(t.mode);
}
console.log(read(new Settings()));

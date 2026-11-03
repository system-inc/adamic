// roads/generic.ts's road handed a class instance, a class assignable to the weak type. The generic's
// one body reads T by its constraint, so the read tests the receiver's storage: the record reads its
// key, and the class instance, neither a fixed object nor a record, stops at the kind test by name,
// natively and on the JavaScript backend. Node reads the instance's field.
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
console.log(read(table));
console.log(read(new Settings()));

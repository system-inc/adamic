// A class instance seen as the weak type is neither a fixed object nor a record, so the read stops by
// name, natively and on the JavaScript backend, and never reads it as a slot. Node reads its field.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["mode"] = 1;
interface Options { readonly total?: number | undefined; readonly mode?: 1 | 2 | undefined }
class Settings {
	total: number | undefined = 3;
	mode: 1 | 2 | undefined = 2;
}
function mode(options: Options): string {
	return String(options.mode);
}
for (const key of Object.keys(table)) console.log(key);
console.log(mode(table));
console.log(mode(new Settings()));

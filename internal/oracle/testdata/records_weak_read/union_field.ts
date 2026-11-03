// moduleNameResolver.ts's shape: packageJsonContent.exports, a union field of a weak type. The field is
// held as a Union, and recordReadDeclared makes no value check for one yet: through the weak type a
// record's values may be of any type, so a number would be read as object | string. The read at 10:16
// stays NotYet until that check exists.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["exports"] = "./record.js";
interface PathFields { readonly total?: number; readonly exports?: object | string; readonly main?: string }
function exported(fields: PathFields): string {
	const entry = fields.exports;
	return typeof entry === "string" ? entry : "none";
}
for (const key of Object.keys(table)) console.log(key);
console.log(exported({ total: 1, exports: "./fixed.js" }));

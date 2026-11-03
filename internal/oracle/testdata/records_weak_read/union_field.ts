// moduleNameResolver.ts's shape: packageJsonContent.exports, a union field of a weak type. A union
// keeps its runtime tag, so the record's value needs no check beyond its storage.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["exports"] = "./record.js";
interface PathFields { total?: number; exports?: object | string; main?: string }
function exported(fields: PathFields): string {
	const entry = fields.exports;
	return typeof entry === "string" ? entry : "none";
}
for (const key of Object.keys(table)) console.log(key);
console.log(exported({ total: 1, exports: "./fixed.js" }));
console.log(exported(table));

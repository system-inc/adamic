// program.ts's shape: result.impliedNodeFormat, a field of the Partial<T> member of a union, read
// once typeof has narrowed the union to it. Fixed objects only, as in optional_plain.ts.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["impliedNodeFormat"] = 99;
interface CreateOptions { readonly total: number; readonly impliedNodeFormat: 1 | 99 }
function format(result: number | Partial<CreateOptions> | undefined): string {
	return typeof result === "object" ? String(result.impliedNodeFormat) : "none";
}
for (const key of Object.keys(table)) console.log(key);
console.log(format({ total: 1, impliedNodeFormat: 1 }));
console.log(format(5));

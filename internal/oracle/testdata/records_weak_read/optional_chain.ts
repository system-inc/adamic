// builder.ts's shape: data?.sourceMapUrlPos, read through ?. on a weak type that may be undefined.
// An undefined receiver reads undefined before its storage is asked about.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["sourceMapUrlPos"] = 40;
interface WriteData { total?: number; sourceMapUrlPos?: number }
function position(data: WriteData | undefined): string {
	return String(data?.sourceMapUrlPos);
}
for (const key of Object.keys(table)) console.log(key);
console.log(position({ total: 1, sourceMapUrlPos: 7 }));
console.log(position(table));
console.log(position(undefined));

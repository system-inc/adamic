type Table = { name: string; count: number; [key: string]: string | number };
const table: Table = { name: "t", count: 1 };
table["extra"] = 5;
table["count"] = 7;
function read<T extends { name: string; count: number }>(t: T): number {
	return t.count;
}
function label<T extends { name: string; count: number }>(t: T): string {
	return t.name;
}
console.log(String(read(table)));
console.log(label(table));

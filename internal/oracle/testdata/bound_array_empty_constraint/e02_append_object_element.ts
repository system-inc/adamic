function append<T extends {}>(to: T[], value: T | undefined): T[];
function append<T extends {}>(to: T[] | undefined, value: T): T[];
function append<T extends {}>(to: T[] | undefined, value: T | undefined): T[] | undefined;
function append<T extends {}>(to: T[] | undefined, value: T | undefined): T[] | undefined {
	if (value === undefined) return to;
	if (to === undefined) return [value];
	to.push(value);
	return to;
}
interface Point {
	a: number;
}
let rows: Point[][] | undefined;
rows = append(rows, [{ a: 1 }]);
rows = append(rows, [{ a: 2 }, { a: 3 }]);
for (const row of rows) {
	let sum = 0;
	for (const point of row) {
		sum += point.a;
	}
	console.log(`${String(row.length)} ${String(sum)}`);
}

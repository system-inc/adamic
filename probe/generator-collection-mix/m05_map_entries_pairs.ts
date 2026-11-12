function* pairs(): Generator<[string, number], void, unknown> { yield ['g', 1]; }
function show(items: Iterable<[string, number]>): void {
	for (const [key, value] of items) console.log(key + '=' + String(value));
}
const table = new Map<string, number>([['a', 10]]);
show(pairs());
show(table.entries());

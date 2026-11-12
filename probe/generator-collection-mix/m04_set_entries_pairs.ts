function* pairs(): Generator<[number, number], void, unknown> { yield [1, 2]; }
function show(items: Iterable<[number, number]>): void {
	for (const [left, right] of items) console.log(String(left) + ':' + String(right));
}
const seen = new Set<number>([7, 8]);
show(pairs());
show(seen.entries());

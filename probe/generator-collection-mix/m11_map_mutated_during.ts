function* counts(): Generator<number, void, unknown> { yield 1; }
function total(values: Iterable<number>, grow: Map<string, number>): number {
	let sum = 0;
	for (const value of values) {
		sum += value;
		if (grow.size < 4) grow.set('k' + String(grow.size), 100);
	}
	return sum;
}
const prices = new Map<string, number>([['a', 10]]);
console.log(String(total(counts(), prices)));
console.log(String(total(prices.values(), prices)));

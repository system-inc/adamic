function* counts(): Generator<number, void, unknown> { yield 1; yield 2; }
function firstTwo(values: IterableIterator<number>): string {
	const a = values.next();
	const b = values.next();
	return String(a.value) + ',' + String(b.value);
}
const prices = new Map<string, number>([['a', 10], ['b', 20]]);
console.log(firstTwo(counts()));
console.log(firstTwo(prices.values()));

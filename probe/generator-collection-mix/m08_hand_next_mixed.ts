function* counts(): Generator<number, void, unknown> { yield 1; yield 2; }
function firstTwo(values: Iterable<number>): string {
	const iterator = values[Symbol.iterator]();
	const a = iterator.next();
	const b = iterator.next();
	return String(a.value) + ',' + String(b.value);
}
const prices = new Map<string, number>([['a', 10], ['b', 20]]);
console.log(firstTwo(counts()));
console.log(firstTwo(prices.values()));

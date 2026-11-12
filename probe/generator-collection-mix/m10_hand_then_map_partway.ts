function* counts(): Generator<number, void, unknown> { yield 1; yield 2; yield 3; }
function firstTwo(values: IterableIterator<number>): string {
	const a = values.next();
	const b = values.next();
	return String(a.value) + ',' + String(b.value);
}
const prices = new Map<string, number>([['a', 10], ['b', 20], ['c', 30]]);
const iterator = prices.values();
iterator.next();
console.log(firstTwo(counts()));
console.log(firstTwo(iterator));

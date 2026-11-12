function* counts(): Generator<number, void, unknown> { yield 1; yield 2; }
function total(values: Iterable<number>): number {
	let sum = 0;
	for (const value of values) sum += value;
	return sum;
}
const prices = new Map<string, number>([['a', 10], ['b', 20], ['c', 30]]);
const iterator = prices.values();
const first = iterator.next();
console.log(String(first.done) + ' ' + String(first.value));
console.log(String(total(counts())));
console.log(String(total(iterator)));

function* low(): Generator<number, void, unknown> { yield 1; }
function* high(): Generator<number, void, unknown> { yield 100; yield 200; }
function total(values: Iterable<number>): number {
	let sum = 0;
	for (const value of values) sum += value;
	return sum;
}
const prices = new Map<string, number>([['a', 10]]);
console.log(String(total(low())));
console.log(String(total(high())));
console.log(String(total(prices.values())));

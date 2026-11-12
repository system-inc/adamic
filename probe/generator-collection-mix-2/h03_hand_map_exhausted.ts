function* counts(): Generator<number, void, unknown> { yield 1; yield 2; yield 3; }
function firstTwo(values: IterableIterator<number>): string {
	let out = "";
	const a = values.next();
	if (!a.done) out += String(a.value);
	const b = values.next();
	if (!b.done) out += "," + String(b.value);
	return out;
}
const prices = new Map<string, number>([["a", 10], ["b", 20], ["c", 30]]);
const iterator = prices.values();
for (const value of iterator) console.log(String(value));
console.log(firstTwo(counts()));
console.log("[" + firstTwo(iterator) + "]");

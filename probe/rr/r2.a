class Lazy {
	get value(): number {
		return 42;
	}
}
interface HasValue {
	readonly value: number;
}
const pairs: [HasValue, number][] = [[new Lazy(), 0]];
for (let [item, index] of pairs) {
	console.log(String(item.value + index));
	item = { value: 1 };
	console.log(String(item.value));
}

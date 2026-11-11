class Lazy {
	get value(): number {
		return 42;
	}
}
interface HasValue {
	readonly value: number;
}
const items: HasValue[] = [new Lazy()];
for (let item of items) {
	console.log(String(item.value));
	item = { value: 1 };
	console.log(String(item.value));
}

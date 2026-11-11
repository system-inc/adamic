class Lazy {
	get value(): number {
		return 42;
	}
}
interface HasValue {
	readonly value: number;
}
const items: HasValue[] = [new Lazy()];
for (const item of items) {
	console.log(String(item.value));
}

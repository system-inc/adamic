class Lazy {
	get value(): number {
		return 42;
	}
}
interface HasValue {
	readonly value: number;
}
function show(item: HasValue): void {
	console.log(String(item.value));
}
function both(first: HasValue, second: HasValue): void {
	console.log(String(first.value + second.value));
}
const head: [HasValue] = [new Lazy()];
both(...head, { value: 1 });

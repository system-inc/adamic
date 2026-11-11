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
const args: [HasValue] = [new Lazy()];
show(...args);

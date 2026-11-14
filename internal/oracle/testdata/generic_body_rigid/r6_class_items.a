class Bag<T extends {}> {
	items: T[] = [];
	widen(): void {
		const wide: {}[] = this.items;
		wide.push(42);
	}
}
function f<T extends {}>(xs: T[]): Bag<T> {
	const bag = new Bag<T>();
	for (const x of xs) {
		bag.items.push(x);
	}
	bag.widen();
	return bag;
}
const bag: Bag<string> = f(["a"]);
for (const item of bag.items) {
	console.log(typeof item);
}

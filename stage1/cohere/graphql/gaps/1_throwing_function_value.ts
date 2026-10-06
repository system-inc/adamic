// Gap 1: a function value that can throw. graphql-js's parser hands each list helper the item's parse
// function, and every parse function can throw.
function total(items: readonly string[], measure: (item: string) => number): number {
	let sum = 0;
	for (const item of items) {
		sum += measure(item);
	}
	return sum;
}
try {
	console.log(`${total(['a', 'bb'], (item) => {
		if (item.length > 1) {
			throw new Error(`too long: ${item}`);
		}
		return item.length;
	})}`);
} catch (error) {
	console.log(error instanceof Error ? error.message : '?');
}

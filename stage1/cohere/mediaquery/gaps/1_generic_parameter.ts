// Closed gap 1: a parameter whose type is the type parameter is monomorphized at this call.
function show<Item>(item: Item, describe: (item: Item) => string): string {
	return describe(item);
}
console.log(show(3, (value) => `${value}`));

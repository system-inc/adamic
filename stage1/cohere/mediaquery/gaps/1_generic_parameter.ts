// Gap 1, from the other side: a parameter whose type is the type parameter is refused too.
function show<Item>(item: Item, describe: (item: Item) => string): string {
	return describe(item);
}
console.log(show(3, (value) => `${value}`));

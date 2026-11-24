const map = new Map<string, number>([['a', 1]]);
function show(iterator: MapIterator<string> & { readonly next: unknown }): void {
	console.log(iterator.toString());
}
show(map.keys());

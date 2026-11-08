// Without a rest: a generic array parameter written through its constraint, then read back as T's element.
function fill<T extends unknown[]>(items: T): T {
	items.push(41);
	return items;
}
const names: string[] = fill(["a"]);
for (const name of names) { console.log(name.toUpperCase()); }

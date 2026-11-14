function append<T extends {}>(to: T[] | undefined, value: T): T[] {
	if (to === undefined) return [value];
	to.push(value);
	return to;
}
interface Item {
	name: string;
}
const items: Item[] = [{ name: "a" }];
const more: Item[] = append(items, { name: "b" });
const fresh: Item[] = append(undefined, { name: "c" });
for (const item of more) {
	console.log(item.name);
}
console.log(String(fresh.length));

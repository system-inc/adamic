// Control: the constraint's push with a value of the element type, no lie.
function grow<T extends string[]>(items: T): T {
	items.push("b");
	return items;
}
const names: string[] = grow(["a"]);
for (const name of names) { console.log(name.toUpperCase()); }

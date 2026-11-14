function append<T extends {}>(to: T[] | undefined, value: T): T[] {
	if (to === undefined) return [value];
	to.push(value);
	return to;
}
interface Cat {
	kind: "cat";
	meow: string;
}
interface Dog {
	kind: "dog";
	bark: string;
}
function choose(dog: boolean): Cat | Dog {
	return dog ? { kind: "dog", bark: "woof" } : { kind: "cat", meow: "mew" };
}
const cats: Cat[] = [{ kind: "cat", meow: "mew" }];
append(cats as Cat[], choose(true));
for (const cat of cats) {
	console.log(typeof cat.meow);
}
